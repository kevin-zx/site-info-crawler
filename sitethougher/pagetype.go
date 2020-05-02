package sitethougher

import (
	"net/url"
	"strings"
)

type PageType string

const (
	PageTypeHome    PageType = "首页"
	PageTypeBread   PageType = "栏目"
	PageTypeNews    PageType = "新闻资讯页"
	PageTypeAbout   PageType = "关于我们"
	PageTypeProduct PageType = "产品页"
	PageTypeANLI    PageType = "案例"
	PageTypeContact PageType = "联系我们"
	PageTypeJoin    PageType = "加入我们"
	PageTypeFAQ     PageType = "常见问题"
	PageTypeKefu    PageType = "客服"
	PageTypeUnKnown PageType = "未识别"
	PageTypeSupport PageType = "技术支持"
)

func setPageType(siteInfo *SiteInfo) {
	//topPageCount := 40
	//if len(siteInfo.SiteLinks) <= 40 {
	//	topPageCount = len(siteInfo.SiteLinks)
	//}
	tpage := len(siteInfo.SiteLinks)
	for i := 1; i < len(siteInfo.SiteLinks); i++ {
		isTop := false
		if siteInfo.SiteLinks[i].QuoteCount >= tpage-10 {
			isTop = true
		}
		if judgeHome(siteInfo.SiteLinks[i].AbsURL) {
			siteInfo.SiteLinks[i].PageType = PageTypeHome
			continue
		}
		if siteInfo.SiteLinks[i].PageType != "" {
			continue
		}
		siteInfo.SiteLinks[i].PageType = judgePageType(siteInfo.SiteLinks[i], siteInfo.Suffix, isTop)
	}

}

func judgeHome(absURL string) bool {
	u, err := url.Parse(absURL)
	if err != nil {
		return false
	}
	if (u.Path == "/" || u.Path == "") && len(u.Query()) == 0 {
		return true
	}
	return false
}

func judgePageType(info *SiteLinkInfo, suffix string, isTopPage bool) PageType {
	urlPts := judgeURL(info.AbsURL)
	t := ""
	if info.WebPageSeoInfo != nil {
		t = info.WebPageSeoInfo.Title
	}
	titlePts := judeTitle(strings.ReplaceAll(t+info.H1+info.HrefTxt, suffix, ""))
	if isTopPage {
		return maxPts(urlPts, titlePts)
	}
	return PageTypeUnKnown
}

func maxPts(pts []PageType, pts2 []PageType) PageType {
	countM := map[PageType]int{}
	for _, pt := range append(pts, pts2...) {
		if _, ok := countM[pt]; ok {
			countM[pt]++
		} else {
			countM[pt] = 1
		}
	}
	p := PageTypeUnKnown
	m := 0
	for pageType, c := range countM {
		if c > m {
			p = pageType
		}
	}
	return p

}

var pageTypeTitleFeatureMap = map[PageType][]string{
	PageTypeContact: {"联系"},
	PageTypeAbout:   {"关于", "简介", "公司介绍", "公司使命", "公司文化", "about", "introduce"},
	PageTypeProduct: {"solution", "产品", "service", "服务"},
	PageTypeJoin:    {"join", "加入", "招聘", "人才", "人力"},
	PageTypeNews:    {"新闻", "资讯"},
	PageTypeANLI:    {"案例"},
	PageTypeFAQ:     {"问题", "知识", "faq", "须知"},
	PageTypeKefu:    {"客服"},
	PageTypeSupport: {"支持"},
}

func judeTitle(title string) []PageType {
	var pts []PageType
	for pageType, features := range pageTypeTitleFeatureMap {
		for _, feature := range features {
			if strings.Contains(title, feature) {
				pts = append(pts, pageType)
				break
			}
		}
	}
	if len(pts) == 0 {
		pts = append(pts, PageTypeUnKnown)
	}
	return pts
}

var pageTypeUrlFeatureMap = map[PageType][]string{
	PageTypeContact: {"lianxi", "contact", "lxwm"},
	PageTypeAbout:   {"about", "guanyu", "gywm", "gongsijianjie", "culture", "jianjie", "introduce"},
	PageTypeProduct: {"solution", "chanpin", "goods", "product", "/cp/", "/cp.", "fuwu"},
	PageTypeJoin:    {"join", "jiaruwomen", "jrwm", "job", "zhaopin", "recruit", "rencai", "rczp", "renliziyuan"},
	PageTypeNews:    {"news", "xinwen", "zixun", "xwzx", "hyxw", "hangyexinwen", "hangye_xinwen", "article", "qyxw", "qiyexinwen"},
	PageTypeANLI:    {"anli", "case"},
	PageTypeFAQ:     {"ask", "faq", "wenti", "zhishi", "xuzhi", "cjwt"},
	PageTypeKefu:    {"kefu"},
	PageTypeSupport: {"jszc", "jishuzhichi", "support"},
}

func judgeURL(absURL string) []PageType {
	var pts []PageType
	absURL = strings.ToLower(absURL)
	for pageType, features := range pageTypeUrlFeatureMap {
		for _, s := range features {
			if strings.Contains(absURL, s) {
				pts = append(pts, pageType)
				break
			}
		}
	}
	if len(pts) == 0 && breadUrlJudge(absURL) {
		pts = append(pts, PageTypeBread)
	}
	return pts
}

var breadUnQueryFeatures = []string{"id", "newsid", "search", "article_id", "pid"}
var breadQueryFeatures = []string{"classid", "cid", "tid", "category_id", "categoryid", "categoryID", "class_id"}
var breadNotFeatures = []string{"page"}
var breadPathFeatures = []string{"list-", "/list/"}
var placeholderPath = []string{"a", "index.php", "html", ""}

func breadUrlJudge(absURL string) bool {
	absURL = strings.ToLower(absURL)
	for _, feature := range breadNotFeatures {
		if strings.Contains(absURL, feature) {
			return false
		}
	}
	if strings.HasSuffix(absURL, "/") {
		return true
	}
	u, err := url.Parse(absURL)
	if err != nil {
		return false
	}
	if len(u.Query()) > 0 {
		for _, b := range breadUnQueryFeatures {
			if _, ok := u.Query()[b]; ok {
				return false
			}
		}
		for _, feature := range breadQueryFeatures {
			if _, ok := u.Query()[feature]; ok {
				return true
			}
		}
		return false
	}
	for _, feature := range breadPathFeatures {
		if strings.Contains(u.Path, feature) {
			return true
		}
	}
	pPart := strings.Split(u.Path, "/")
	pCount := 0
	for _, s := range pPart {
		if strArrayIn(placeholderPath, s) {
			continue
		}
		pCount++
	}
	if pCount == 1 {
		return true
	}
	return false
}
func isMatchFeature() {

}

func mean(nums []float64) float64 {
	var t float64 = 0
	for _, num := range nums {
		t += num
	}
	return t / float64(len(nums))
}

func strArrayIn(strArr []string, instr string) bool {
	for _, s := range strArr {
		if s == instr {
			return true
		}
	}
	return false
}
