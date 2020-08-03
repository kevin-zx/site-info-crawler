package sitethrougher

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

type GTSOption struct {
	LimitCount    int
	Delay         time.Duration
	Port          DevicePort
	TimeOut       time.Duration
	AllowedDomain string
	//NeedDocument bool
}

const fileRegString = ".+?(\\.jpg|\\.png|\\.gif|\\.GIF|\\.PNG|\\.JPG|\\.pdf|\\.PDF|\\.doc|\\.DOC|\\.csv|\\.CSV|\\.xls|\\.XLS|\\.xlsx|\\.XLSX|\\.mp40|\\.lfu|\\.DNG|\\.ZIP|\\.zip)(\\W+?\\w|$)"

var tmpLock = new(sync.Mutex)

func goThoughtSite(siteUrlStr string, gtsOption *GTSOption,
	handler func(html *colly.HTMLElement),
	onErr func(response *colly.Response, e error),
	parentInfo func(currentUrl string, parentUrl string, keyword string)) (err error) {
	//userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"

	userAgent := "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)"
	if gtsOption.Port == PortMobile {
		userAgent = "Mozilla/5.0 (Linux;u;Android 4.2.2;zh-cn;) AppleWebKit/534.46 (KHTML,like Gecko) Version/5.1 Mobile Safari/10600.6.3 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)"
	}
	siteUrl, err := url.Parse(siteUrlStr)
	if err != nil {
		return err
	}
	if gtsOption.AllowedDomain == "" {
		gtsOption.AllowedDomain = siteUrl.Host
	}

	c := colly.NewCollector(
		colly.AllowedDomains(gtsOption.AllowedDomain, siteUrl.Host),
		colly.DisallowedURLFilters(regexp.MustCompile(fileRegString)),
		colly.UserAgent(userAgent),
		colly.Async(true),
		colly.MaxDepth(1000),
	)
	requestMap := make(map[string]int)

	c.DetectCharset = true
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*" + siteUrl.Host + "*",
		Parallelism: 4,
		RandomDelay: gtsOption.Delay,
		Delay:       gtsOption.Delay,
	})
	if err != nil {
		return err
	}
	c.SetRequestTimeout(gtsOption.TimeOut)
	c.OnHTML("html", func(ele *colly.HTMLElement) {
		if ele.Request.ID%50 == 0 {
			log.Printf("爬取了 %d 个\n", ele.Request.ID)
		}
		if handler != nil {
			handler(ele)
		}

		ele.DOM.Find("a[href]").EachWithBreak(func(i int, a *goquery.Selection) bool {
			href, ok := a.Attr("href")
			if !ok {
				return true
			}
			if strings.Contains(href, "script") {
				return true
			}

			link := clearUrl(href)

			resultUrl := parseUrl(ele.Request.URL, link)
			hrefText := a.Text()
			if hrefText == "" {
				hrefText = a.AttrOr("alt", "")
			}
			if resultUrl != "" && strings.HasPrefix(resultUrl, "http") {

				//_ = ele.Request.Visit(resultUrl)
				flag := true
				if len(c.DisallowedURLFilters) > 0 {
					if isMatchingFilter(c.DisallowedURLFilters, []byte(resultUrl)) {
						flag = false
					}
				}
				if len(c.URLFilters) > 0 {
					if !isMatchingFilter(c.URLFilters, []byte(resultUrl)) {
						flag = false
					}
				}
				if myFilter(resultUrl, siteUrl.Host) {
					flag = false
				}
				if flag {
					parentInfo(resultUrl, ele.Request.URL.String(), hrefText)
					_ = ele.Request.Visit(resultUrl)
				}
			}
			return true
		})
	})

	c.OnRequest(func(request *colly.Request) {
		tmpLock.Lock()
		defer tmpLock.Unlock()
		if len(requestMap) > gtsOption.LimitCount {
			request.Abort()
		}

		requestMap[request.URL.String()] = 0

	})
	c.OnResponse(func(response *colly.Response) {
		//fmt.Println("-------",response.Request.URL.String())
		if !strings.Contains(strings.ToLower(response.Headers.Get("Content-Type")), "html") {
			response.Headers.Set("Content-Type", "text/html;")
		}

	})
	c.OnError(func(response *colly.Response, err error) {
		onErr(response, err)
	})
	err = c.Visit(siteUrl.String())
	c.Wait()
	return err
}

func parseUrl(curl *url.URL, href string) string {
	testUrl, _ := curl.Parse(href)
	testUrlStr := ""
	if testUrl != nil {
		testUrlStr = testUrl.String()
	}
	return testUrlStr
}

func clearUrl(webUrl string) string {

	webUrl = handlerDoubleSlant(webUrl)
	//去除空格
	webUrl = strings.TrimSpace(webUrl)
	//utf-8空格
	for strings.HasSuffix(webUrl, "%20") {
		webUrl = string(webUrl[0 : len(webUrl)-3])
	}
	//unicode空格
	webUrl = strings.Replace(webUrl, "&#10;", "", -1)
	webUrl = strings.Replace(webUrl, "&#9;", "", -1)
	return webUrl

}

func handlerDoubleSlant(webUrl string) string {
	for strings.HasSuffix(webUrl, "//") {
		webUrl = strings.Replace(webUrl, "//", "/", -1)
	}
	return webUrl
}
func isMatchingFilter(fs []*regexp.Regexp, d []byte) bool {
	for _, r := range fs {
		if r.Match(d) {
			return true
		}
	}
	return false
}

var notAllowedFeatures = []string{"#"}

func myFilter(url string, host string) bool {
	if !strings.Contains(url, host) {
		return true
	}
	for _, feature := range notAllowedFeatures {
		if strings.Contains(url, feature) {
			return true
		}
	}
	return false
}
