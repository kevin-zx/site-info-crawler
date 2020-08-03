package sitethrougher

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"log"
	"net/url"
	"regexp"
	"strings"
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
		colly.AllowedDomains(gtsOption.AllowedDomain,siteUrl.Host),
		colly.DisallowedURLFilters(regexp.MustCompile(fileRegString)),
		colly.UserAgent(userAgent),
		colly.Async(true),
		colly.MaxDepth(1000),
	)

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

		ele.DOM.Find("a[href]").Each(func(i int, a *goquery.Selection) {
			href, ok := a.Attr("href")
			if !ok {
				return
			}
			if strings.Contains(href, "script") {
				return
			}
			link := clearUrl(href)

			resultUrl := parseUrl(ele.Request.URL, link)
			hrefText := a.Text()
			if hrefText == "" {
				hrefText = a.AttrOr("alt", "")
			}
			if resultUrl != "" && strings.HasPrefix(resultUrl, "http") {
				parentInfo(resultUrl, ele.Request.URL.String(), hrefText)
				_ = ele.Request.Visit(resultUrl)
			}
		})
	})
	c.OnRequest(func(request *colly.Request) {
		if request.ID > uint32(gtsOption.LimitCount) {
			request.Abort()
		}else{
			//fmt.Println(request.URL.String())
		}
	})
	c.OnResponse(func(response *colly.Response) {
		//fmt.Println("-------",response.Request.URL.String())
		if !strings.Contains(strings.ToLower(response.Headers.Get("Content-Type")), "html") {
			response.Headers.Set("Content-Type", "text/html;")
		}

	})
	c.OnError(onErr)
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
