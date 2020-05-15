package sitethougher

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

type WebPageSeoInfo struct {
	Title       string
	Description string
	Keywords    string
	RealUrl     *url.URL
	RecordCount int
	InitUrl     string
	Url         url.URL
}

type SiteLinkInfo struct {
	AbsURL         string
	StatusCode     int
	ParentURL      string
	Depth          int
	WebPageSeoInfo *WebPageSeoInfo
	H1             string
	IsCrawler      bool
	InnerText      string
	HrefTxt        string
	QuoteCount     int // 引用次数
	PageType       PageType
}

type SiteInfo struct {
	SiteLinks []*SiteLinkInfo
	Suffix    string
}

var splitText = []string{",", "-", "，", "、", "_", " ", "\t", ";", "；", "\n", "“", "”", "\""}

func (wpsi *WebPageSeoInfo) SpiltKeywordsStr2Arr() (keywords []string) {
	// 处理keywordStr 到arr
	keywordsStr := wpsi.Keywords
	//替换统一的分隔符
	for _, s := range splitText {
		keywordsStr = strings.Replace(keywordsStr, s, "|", -1)
	}
	keywords = removeDuplicatesAndEmpty(strings.Split(keywordsStr, "|"))
	return
}

type DevicePort int

const (
	PortPC     DevicePort = 1
	PortMobile DevicePort = 2
)

func RunWithParams(siteUrlRaw string, limitCount int, delay time.Duration, port DevicePort) (si *SiteInfo, err error) {
	// 这里的锁一定不能暴露到方法外部不然就线程不安全了
	mu := sync.Mutex{}
	linkMap := map[string]*SiteLinkInfo{siteUrlRaw: {AbsURL: siteUrlRaw}}
	err = goThoughtSite(siteUrlRaw, port, limitCount, delay, func(html *colly.HTMLElement) {
		currentUrl := html.Request.URL.String()

		wi, err := parseWebSeoElement(html.DOM)
		if err != nil {
			//panic(err)
			return
		}
		si := linkMap[currentUrl]
		// todo: 了解这里的原因，然后更加精细的处理
		if si == nil {
			return
		}
		h1 := html.DOM.Find("h1")
		mu.Lock()
		if html.DOM.Find("body") != nil {
			si.InnerText = html.DOM.Find("body").Text()
		}

		html.DOM.Find("script").Each(func(_ int, selection *goquery.Selection) {
			si.InnerText = strings.Replace(si.InnerText, selection.Text(), "", -1)
		})
		TextLen := len(strings.Split(si.InnerText, ""))
		if TextLen > 8000 {
			TextLen = 8000
		}
		si.InnerText = strings.Join(strings.Split(si.InnerText, "")[0:TextLen], "")
		si.IsCrawler = true
		si.H1 = clear(h1.Text())
		si.WebPageSeoInfo = wi
		si.Depth = html.Request.Depth
		if html.Response.StatusCode != 200 {
			fmt.Println(html.Response.StatusCode)
		}
		si.StatusCode = html.Response.StatusCode
		mu.Unlock()
	}, func(response *colly.Response, e error) {
		mu.Lock()
		currentUrl := response.Request.URL.String()
		if !linkMap[currentUrl].IsCrawler {
			linkMap[currentUrl].IsCrawler = true
			linkMap[currentUrl].Depth = response.Request.Depth
			linkMap[currentUrl].StatusCode = response.StatusCode
		}
		mu.Unlock()
	}, func(currentUrl string, parentUrl string, hrefTxt string, err error) {
		mu.Lock()
		if _, ok := linkMap[currentUrl]; !ok {
			linkMap[currentUrl] = &SiteLinkInfo{AbsURL: currentUrl, HrefTxt: clear(hrefTxt), ParentURL: parentUrl}
		}

		linkMap[currentUrl].QuoteCount = linkMap[currentUrl].QuoteCount + 1
		mu.Unlock()
		return
	})
	si = &SiteInfo{SiteLinks: []*SiteLinkInfo{}, Suffix: ""}
	var ts []string
	for k, v := range linkMap {
		if !v.IsCrawler {
			delete(linkMap, k)
			continue
		}
		if v.ParentURL == "" {
			v.PageType = PageTypeHome
		}
		if v.WebPageSeoInfo != nil && v.WebPageSeoInfo.Title != "" {
			ts = append(ts, v.WebPageSeoInfo.Title)
		}
		si.SiteLinks = append(si.SiteLinks, v)
	}

	// 这里降序排序
	sort.Slice(si.SiteLinks, func(i, j int) bool {
		return si.SiteLinks[i].QuoteCount > si.SiteLinks[j].QuoteCount
	})
	ps := GetPublicSuffix(ts)
	si.Suffix = ps
	setPageType(si)
	return
}

func convertGBKCharset(sli *SiteLinkInfo) {
	h1B, err := gbkToUtf8([]byte(sli.H1))
	if err == nil {
		sli.H1 = string(h1B)
	}
	innerTextByte, err := gbkToUtf8([]byte(sli.InnerText))
	if err == nil {
		fmt.Println(sli.InnerText)
		sli.InnerText = string(innerTextByte)
		fmt.Println(sli.InnerText)
	}
	descByte, err := gbkToUtf8([]byte(sli.WebPageSeoInfo.Description))
	if err == nil {
		sli.WebPageSeoInfo.Description = string(descByte)
	}
	keywordsByte, err := gbkToUtf8([]byte(sli.WebPageSeoInfo.Keywords))
	if err == nil {
		sli.WebPageSeoInfo.Keywords = string(keywordsByte)
	}

	TitleByte, err := gbkToUtf8([]byte(sli.WebPageSeoInfo.Title))
	if err == nil {
		sli.WebPageSeoInfo.Title = string(TitleByte)
	}

}

func gbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GB18030.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
func clear(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, " ")
	s = strings.Replace(s, "​", "", -1)
	s = strings.Replace(s, " ", "", -1)
	s = strings.Replace(s, " ", "", -1)
	s = strings.Replace(s, "	", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	return s
}

func parseWebSeoElement(html *goquery.Selection) (*WebPageSeoInfo, error) {
	title := html.Find("title").Text()
	description, _ := html.Find("meta[name=description]").Attr("content")
	keywords, _ := html.Find("meta[name=keywords]").Attr("content")
	site := WebPageSeoInfo{Title: title, Description: description, Keywords: keywords}
	return &site, nil
}

func removeDuplicatesAndEmpty(a []string) (ret []string) {
	var keywordCount = make(map[string]int)
	aLen := len(a)
	for i := 0; i < aLen; i++ {
		duFlag := false
		for _, re := range ret {
			if len(a[i]) == 0 {
				duFlag = true
				break
			}
			if re == a[i] {
				if _, ok := keywordCount[re]; !ok {
					keywordCount[re] = 1
				}
				duFlag = true
				break
			}
		}
		if !duFlag {
			ret = append(ret, a[i])
		}
	}
	return
}
