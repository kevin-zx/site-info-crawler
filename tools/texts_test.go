package tools

import (
	"fmt"
	site_info_crawler "github.com/kevin-zx/site-info-crawler"
	"github.com/kevin-zx/site-info-crawler/sitethrougher"
	"testing"
)

func TestDiscriminateSiteTextSamePart(t *testing.T) {
	o := sitethrougher.DefaultOption
	o.NeedDocument = true
	si, err := site_info_crawler.RunSiteWithCache("http://www.haimicloud.com/", "../data/", 10000, o)
	if err != nil {
		panic(err)
	}

	sps := DiscriminateSiteTextSamePart(si)
	for _, sp := range sps {
		fmt.Println(len(sp.Text), sp.Rate)
	}
}
