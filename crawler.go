package site_info_crawler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/kevin-zx/site-info-crawler/sitethrougher"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func RunSite(siteURL string) (si *sitethrougher.SiteInfo, err error) {
	return sitethrougher.RunWithOptions(siteURL, nil)
}

// RunSiteWithCache: go through the site and will write to the cache file, if between a recent hour
// to now already ran a loop, just read this cache file instead run a new loop.
// siteURL: must contain protocol.
// cachePath: specify a cache path to find and save cache file.
// expireHour: if in now - expireHour to now has cache file history, will read and decode file to result.
// option: go through option.
func RunSiteWithCache(siteURL string, cachePath string, expireHour int, option *sitethrougher.Option) (si *sitethrougher.SiteInfo, err error) {
	pu, err := url.Parse(siteURL)
	if err != nil {
		return nil, err
	}
	if pu.Host == "" {
		return nil, fmt.Errorf("ilegal url:%s", siteURL)
	}
	cacheFile, _ := findRecentCache(cachePath, pu.Host, option.LimitCount, int(option.Port), option.NeedDocument, expireHour)
	if cacheFile != "" {
		si, err = decodeJsonFileToSiteInfo(cacheFile, cachePath)
		if err == nil {
			return
		}
	}

	fileName := encodeFileName(pu.Host, option.LimitCount, int(option.Port), option.NeedDocument, time.Now().Unix())
	cacheFileName := cachePath + "/" + fileName

	si, err = sitethrougher.RunWithOptions(siteURL, option)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(si)
	if err != nil {
		return nil, err
	}
	f, err := os.Create(cacheFileName)
	if err != nil {
		return
	}
	defer f.Close()
	bufW := bufio.NewWriter(f)
	_, err = bufW.Write(data)
	bufW.Flush()
	return si, err
}

// decode data to JSON and write to cache file
func decodeJsonFileToSiteInfo(fileName string, cachePath string) (si *sitethrougher.SiteInfo, err error) {
	data, err := ioutil.ReadFile(cachePath + "/" + fileName)
	if err == nil {
		si := &sitethrougher.SiteInfo{}
		if len(data) > 0 {
			err = json.Unmarshal(data, si)
			return si, err
		}
	}
	return
}

// generate fileName by domain limitCount ... , those params can't ensure the unique cache condition
func encodeFileName(domain string, limitCount int, port int, needDocument bool, timeStamp int64, ) string {
	return fmt.Sprintf("sitebak_%s_%d_%d_%v_%d", domain, limitCount, port, needDocument, timeStamp) + ".json"
}

// decode file name to domain, limitCount, port ...
func decodeFileName(name string) (domain string, limitCount int, port int, needDocument bool, timeStamp int64, err error) {
	parts := strings.Split(strings.ReplaceAll(name, ".json", ""), "_")
	if len(parts) == 5 {
		domain = parts[0]
		limitCount, _ = strconv.Atoi(parts[1])
		port, _ = strconv.Atoi(parts[2])
		needDocument = parts[3] == "true"
		timeStamp, _ = strconv.ParseInt(parts[4], 10, 64)
		if limitCount == 0 || port == 0 || timeStamp == 0 {
			err = fmt.Errorf("this file name is't cache file")
			return
		}
	}
	return
}

// find recent cache file by unique conditions.
func findRecentCache(cachePath string, domain string, limitCount int, port int, needDocument bool, recentHour int) (string, error) {
	fs, err := ioutil.ReadDir(cachePath)
	if err != nil {
		return "", err
	}
	nowTimeStamp := time.Now().Unix()
	timeLimit := int64(recentHour) * 3600
	for _, f := range fs {
		if !f.IsDir() && strings.Contains(f.Name(), "sitebak_") {
			cDomain, cLimit, cPort, cNeed, cTimeStamp, err := decodeFileName(f.Name())
			if err != nil {
				continue
			}
			if domain == cDomain && limitCount <= cLimit && cPort == port && cNeed == needDocument && nowTimeStamp-cTimeStamp <= timeLimit {
				return f.Name(), nil
			}
		}
	}
	return "", fmt.Errorf("can't find rencet cache file by domain:%s", domain)
}
