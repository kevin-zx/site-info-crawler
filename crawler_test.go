package site_info_crawler

import (
	"encoding/json"
	"fmt"
	"github.com/kevin-zx/site-info-crawler/sitethrougher"
	"os"
)

func ExampleRunSiteWithCache() {
	si, err := RunSiteWithCache("http://www.jianshu.com", os.TempDir(), 24*1, sitethrougher.DefaultOption)
	if err != nil {
		panic(err)
	}
	fmt.Printf("suffix:%s", si.Suffix)
	for _, link := range si.SiteLinks {
		data, _ := json.Marshal(link)
		fmt.Sprintln(string(data))
	}
}
