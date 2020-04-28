package gothoughtsite

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/url"
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
	QuoteCount     int // 引用次数
}

//var mu sync.Mutex

func (wpsi *WebPageSeoInfo) SpiltKeywordsStr2Arr() (keywords []string) {
	// 处理keywordStr 到arr
	keywordsStr := wpsi.Keywords
	//替换统一的分隔符
	keywordsStr = strings.Replace(keywordsStr, ",", "|", -1)
	keywordsStr = strings.Replace(keywordsStr, "-", "|", -1)
	keywordsStr = strings.Replace(keywordsStr, "，", "|", -1)
	keywordsStr = strings.Replace(keywordsStr, "、", "|", -1)
	keywordsStr = strings.Replace(keywordsStr, "_", "|", -1)
	keywordsStr = strings.Replace(keywordsStr, " ", "|", -1)
	keywordsStr = strings.Replace(keywordsStr, "\t", "|", -1)
	keywordsStr = strings.Replace(keywordsStr, ";", "|", -1)
	keywordsStr = strings.Replace(keywordsStr, "；", "|", -1)

	keywordsStr = strings.Replace(keywordsStr, "\n", "", -1)
	keywordsStr = strings.Replace(keywordsStr, "“", "", -1)
	keywordsStr = strings.Replace(keywordsStr, "”", "", -1)
	keywords = removeDuplicatesAndEmpty(strings.Split(keywordsStr, "|"))
	return
}

type DevicePort int

const (
	PortPC     DevicePort = 1
	PortMobile DevicePort = 2
)

func RunWithParams(siteUrlRaw string, limitCount int, delay time.Duration, port DevicePort) (linkMap map[string]*SiteLinkInfo, err error) {
	// 这里的锁一定不能暴露到方法外部不然就线程不安全了
	mu := sync.Mutex{}
	linkMap = map[string]*SiteLinkInfo{siteUrlRaw: {}}
	err = goThoughtSite(siteUrlRaw, port, limitCount, delay, func(html *colly.HTMLElement) {
		wi, err := parseWebSeoElement(html.DOM)
		if err != nil {
			return
		}

		currentUrl := html.Request.URL.String()

		h1 := html.DOM.Find("h1")
		mu.Lock()
		if _, ok := linkMap[currentUrl]; !ok {
			linkMap[currentUrl] = &SiteLinkInfo{AbsURL: currentUrl}
		}

		linkMap[currentUrl].InnerText = html.DOM.Find("body").Text()
		//fmt.Println(linkMap[currentUrl].InnerText)
		TextLen := len(strings.Split(linkMap[currentUrl].InnerText, ""))
		if TextLen > 8000 {
			TextLen = 8000
		}

		linkMap[currentUrl].InnerText = strings.Join(strings.Split(linkMap[currentUrl].InnerText, "")[0:TextLen], "")
		//fmt.Println(linkMap[currentUrl].InnerText )
		linkMap[currentUrl].IsCrawler = true
		linkMap[currentUrl].H1 = clear(h1.Text())
		linkMap[currentUrl].WebPageSeoInfo = wi
		//charset,_ := html.DOM.Find("meta[http-equiv=Content-Type]").Attr("content")
		//if strings.Contains(strings.ToLower(charset),"gb") {
		//	convertGBKCharset(linkMap[currentUrl])
		//}
		linkMap[currentUrl].Depth = html.Request.Depth
		if html.Response.StatusCode != 200 {
			fmt.Println(html.Response.StatusCode)
		}
		linkMap[currentUrl].StatusCode = html.Response.StatusCode
		mu.Unlock()
	}, func(response *colly.Response, e error) {
		mu.Lock()
		currentUrl := response.Request.URL.String()
		if _, ok := linkMap[currentUrl]; !ok {
			linkMap[currentUrl] = &SiteLinkInfo{AbsURL: currentUrl}
		}
		existLink := linkMap[currentUrl]
		fmt.Println(existLink.StatusCode)
		if !linkMap[currentUrl].IsCrawler {
			linkMap[currentUrl].IsCrawler = true
			linkMap[currentUrl].Depth = response.Request.Depth
			linkMap[currentUrl].StatusCode = response.StatusCode
		}

		mu.Unlock()
	}, func(currentUrl string, parentUrl string, err error) {
		mu.Lock()
		if _, ok := linkMap[currentUrl]; !ok {
			linkMap[currentUrl] = &SiteLinkInfo{AbsURL: currentUrl}
		}
		linkMap[currentUrl].QuoteCount = linkMap[currentUrl].QuoteCount + 1
		linkMap[currentUrl].ParentURL = parentUrl
		if err != nil && err.Error() == "URL already visited" {
			linkMap[currentUrl].QuoteCount = linkMap[currentUrl].QuoteCount + 1
		}
		mu.Unlock()
		return
	})
	mu.Lock()
	for k, v := range linkMap {
		// todo:会有absUrl为空的情况，暂时搞不懂为什么。先暴力修复
		if v.AbsURL == "" {
			linkMap[k].AbsURL = k
		}
		if !v.IsCrawler {
			delete(linkMap, k)
		}
	}
	mu.Unlock()
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
