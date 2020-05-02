package sitethougher

import (
	"fmt"
	"testing"
	"time"
)

func TestRunWithParams(t *testing.T) {
	//
	////gotLinkMap, err := RunWithParams("http://www.whbfyf.com/", 10000,
	gotLinkMap, err := RunWithParams("http://www.tpsyyq.com/", 30,
		time.Second*1, PortPC)
	if err != nil {
		panic(err)
	}
	fmt.Println(gotLinkMap.Suffix)

	//
	for _, l := range gotLinkMap.SiteLinks {
		title := ""
		if l.WebPageSeoInfo != nil {
			title = l.WebPageSeoInfo.Title
		}
		fmt.Printf("aburl: %s , title: %s , quoteCount: %d , innTextLen: %d, hrefText: %s ， pageType:%s\n", l.AbsURL, title, l.QuoteCount, len(l.InnerText), l.HrefTxt, l.PageType)
	}
}
