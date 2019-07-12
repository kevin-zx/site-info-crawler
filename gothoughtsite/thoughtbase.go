package gothoughtsite

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const fileRegString = ".+?(\\.jpg|\\.png|\\.gif|\\.GIF|\\.PNG|\\.JPG|\\.pdf|\\.PDF|\\.doc|\\.DOC|\\.csv|\\.CSV|\\.xls|\\.XLS|\\.xlsx|\\.XLSX|\\.mp40|\\.lfu|\\.DNG|\\.ZIP|\\.zip)(\\W+?\\w|$)"

func goThoughtSite(siteUrlStr string, port int, limitCount int, timeOut time.Duration,
	handler func(html *colly.HTMLElement),
	onErr func(response *colly.Response, e error),
	parentInfo func(currentUrl string, parentUrl string, err error)) (err error) {
	//userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"

	userAgent := "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)"
	if port == 2 {
		userAgent = "Mozilla/5.0 (Linux;u;Android 4.2.2;zh-cn;) AppleWebKit/534.46 (KHTML,like Gecko) Version/5.1 Mobile Safari/10600.6.3 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html）"
	}
	siteUrl, err := url.Parse(siteUrlStr)
	if err != nil {
		return err
	}

	c := colly.NewCollector(
		colly.AllowedDomains(siteUrl.Host),
		colly.DisallowedURLFilters(regexp.MustCompile(fileRegString)),
		colly.UserAgent(userAgent),
		colly.Async(true),
		colly.MaxDepth(1000),
	)
	c.DetectCharset = true
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*" + siteUrl.Host + "*",
		Parallelism: 4,
		RandomDelay: timeOut,
		Delay:       timeOut,
	})
	c.SetRequestTimeout(20 * time.Second)
	c.OnHTML("html", func(ele *colly.HTMLElement) {

		fmt.Println(ele.Request.ID)
		if ele.Request.ID%50 == 0 {
			fmt.Printf("爬取了 %d 个\n", ele.Request.ID)
		}
		if handler != nil {
			handler(ele)
		}

		ele.DOM.Find("a[href]").Each(func(i int, a *goquery.Selection) {
			href, ok := a.Attr("href")
			if !ok {
				return
			}
			link := clearUrl(href)
			resultUrl := parseUrl(ele.Request.URL, link)

			//if resultUrl != "" && (strings.HasPrefix(resultUrl, "http:") || strings.HasPrefix(resultUrl, "https:")) && !strings.HasPrefix(link, "./") && !strings.HasPrefix(link, "/") && !strings.HasPrefix(link, "http") {
			//	link = "/" + link
			//	resultUrl = parseUrl(ele.Request.URL, link)
			//}
			if resultUrl != "" {
				erri := ele.Request.Visit(resultUrl)
				parentInfo(resultUrl, ele.Request.URL.String(), erri)
				//if erri == nil {
				//
				//}else{
				//	fmt.Println(erri.Error())
				//}
			}
		})
	})
	c.OnRequest(func(request *colly.Request) {
		if request.ID > uint32(limitCount) {
			request.Abort()
		}
		//request.Headers.Add("Accept-Encoding","gzip, deflate")
		//
		//request.Headers.Add("Accept","text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
		//request.Headers.Add("Accept-Encoding","gzip, deflate")
		//request.Headers.Add("Accept-Language","en,zh-CN;q=0.9,zh;q=0.8,en-US;q=0.7,zh-TW;q=0.6")
		//request.Headers.Add("Connection","keep-alive")
		//request.Headers.Add("Host","www.oubeisiman.com")
		//request.Headers.Add("Referer","http://www.oubeisiman.com/pd.jsp?id=133")
		//request.Headers.Add("Upgrade-Insecure-Requests","1")
	})
	c.OnResponse(func(response *colly.Response) {
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
	//if !strings.HasPrefix(webUrl,"./") && !strings.HasPrefix(webUrl,"/") && !strings.HasPrefix(webUrl,"http"){
	//	webUrl = "/"+webUrl
	//}
	return webUrl

}

func handlerDoubleSlant(webUrl string) string {
	for strings.HasSuffix(webUrl, "//") {
		webUrl = strings.Replace(webUrl, "//", "/", -1)
	}
	return webUrl
}
