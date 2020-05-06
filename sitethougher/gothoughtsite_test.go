package sitethougher

import (
	"fmt"
	"testing"
	"time"
)

func TestRunWithParams(t *testing.T) {
	//
	////si, err := RunWithParams("http://www.whbfyf.com/", 10000,
	si, err := RunWithParams("http://www.zggsjx.com/", 3000,
		time.Second*1, PortPC)
	if err != nil {
		panic(err)
	}
	fmt.Println(si.Suffix)
	//result, err := json.Marshal(si)
	//if err != nil {
	//	panic(err)
	//}
	//f, err := os.Create("../data/test/si.data")
	//if err != nil {
	//	panic(err)
	//}
	//defer f.Close()
	//bufw := bufio.NewWriter(f)
	//_, err = bufw.Write(result)
	//if err != nil {
	//	panic(err)
	//}
	//bufw.Flush()
	//
	for _, l := range si.SiteLinks {
		title := ""
		if l.WebPageSeoInfo != nil {
			title = l.WebPageSeoInfo.Title
		}
		fmt.Printf("aburl: %s , title: %s , quoteCount: %d , innTextLen: %d, hrefText: %s ， pageType:%s\n", l.AbsURL, title, l.QuoteCount, len(l.InnerText), l.HrefTxt, l.PageType)
	}
}
