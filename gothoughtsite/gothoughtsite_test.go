package gothoughtsite

import (
	"fmt"
	"testing"
	"time"
)

func TestRunWithParams(t *testing.T) {

	//gotLinkMap, err := RunWithParams("http://www.whbfyf.com/", 10000,
	gotLinkMap, err := RunWithParams("http://www.sichuanks.cn/", 10000,
		time.Second*1, 1)
	if err != nil {
		panic(err)
	}

	for _, l := range gotLinkMap {
		title := ""
		if l.WebPageSeoInfo != nil {
			title = l.WebPageSeoInfo.Title
		}
		fmt.Printf("%s , %s , %d , %d\n", l.AbsURL, title, l.QuoteCount, len(l.InnerText))
	}
}
