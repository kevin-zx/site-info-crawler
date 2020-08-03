package urlgroup

import (
	"net/url"
	"sort"
	"strings"
)

type URLGroup struct {
	Deep          int
	URLFeature    *URLFeature
	NextURLGroups []*URLGroup
}

func (ug *URLGroup) Insert(feature *URLFeature) {
	if len(feature.PathFeature) == 0 && len(feature.QueryFeature) == 0 {
		return
	}
	for _, group := range ug.NextURLGroups {
		if isSubFeature(*group.URLFeature, *feature) {
			group.Insert(feature)
			return
		}
	}

	fPathlen := len(feature.PathFeature)
	ugPathLen := len(ug.URLFeature.PathFeature)
	if fPathlen > ugPathLen+1 {
		unRealFeatures := genMidURLByPath(feature, ug.URLFeature)
		for _, unRealFeature := range unRealFeatures {
			ug.Insert(&unRealFeature)
		}
		ug.Insert(feature)
		return
	}
	//fmt.Println(strings.Join(ug.URLFeature.PathFeature, "/"), len(ug.NextURLGroups))
	ng := newGroupByFeature(*feature, ug.Deep)
	ug.NextURLGroups = append(ug.NextURLGroups, &ng)
	return
}

func genMidURLByPath(feature *URLFeature, feature2 *URLFeature) []URLFeature {
	pu, _ := url.Parse(feature.URLRaw)
	base := pu.Scheme + "://" + pu.Host + "/"
	var ufs []URLFeature
	for i := range feature.PathFeature {
		if i > len(feature2.PathFeature)-1 {
			gu := base + strings.Join(feature.PathFeature[:i+1], "/")
			uf, err := decomposeSingleURL(gu, false)
			if err != nil {
				panic(err)
			}
			ufs = append(ufs, *uf)
		}
	}
	base = pu.Scheme + "://" + pu.Host + "/" + pu.Path
	if len(feature.QueryFeature) > len(feature2.QueryFeature)+1 {
		for i := range feature.QueryFeature {
			if i > len(feature2.QueryFeature)-1 {

				gu := base + "?" + convertQVs2rawQuery(feature.QueryFeature[:i+1])
				uf, err := decomposeSingleURL(gu, false)
				if err != nil {
					panic(err)
				}
				ufs = append(ufs, *uf)
			}
		}
	}
	return ufs
}

func convertQVs2rawQuery(qvs []QV) string {
	qs := []string{}
	for _, qv := range qvs {
		qs = append(qs, qv.K+"="+strings.Join(qv.V, ","))
	}
	return strings.Join(qs, "&")
}

func newGroupByFeature(feature URLFeature, deep int) URLGroup {
	ng := URLGroup{
		Deep: deep + 1,
		//TopURLRaw:      feature.URLRaw,
		URLFeature: &feature,
		//PathFeature:    "",
		//QueryFeature:   nil,
		//URLs:           nil,
		NextURLGroups: []*URLGroup{},
	}
	return ng
}

func (ug *URLGroup) isSub(feature *URLFeature) bool {
	if len(ug.URLFeature.PathFeature) > len(feature.PathFeature) {
		return false
	}
	for i, s := range ug.URLFeature.PathFeature {
		if strings.ReplaceAll(s, ".html", "") != strings.ReplaceAll(feature.PathFeature[i], ".html", "") {
			return false
		}
	}
	if len(ug.URLFeature.PathFeature) < len(feature.PathFeature) {
		return true
	}

	if len(ug.URLFeature.QueryFeature) > len(feature.QueryFeature) {
		return false
	}
	for i, s := range ug.URLFeature.QueryFeature {
		if s.K != feature.QueryFeature[i].K {
			return false
		} else {
			if len(s.V) != len(feature.QueryFeature[i].V) {
				return false
			}
			for vi, vs := range s.V {
				if vs != feature.QueryFeature[i].V[vi] {
					return false
				}
			}
		}
	}

	return true
}

type QV struct {
	K string
	V []string
}

type URLFeature struct {
	URLRaw       string
	PathFeature  []string
	QueryFeature []QV
	Exist        bool
}

func GroupURLs(urls []string) URLGroup {
	ufs := DecomposeURLs(urls)
	ufs = URLFeatureSort(ufs)
	u, err := url.Parse(urls[0])
	if err != nil {
		panic(err)
	}

	ug := URLGroup{
		//TopURLRaw: u.Scheme + "://" + u.Host,
		URLFeature: &URLFeature{
			URLRaw:       u.Scheme + "://" + u.Host,
			PathFeature:  []string{},
			QueryFeature: nil,
		},

		NextURLGroups: []*URLGroup{},
	}

	for i := len(ufs) - 1; i >= 0; i-- {
		ug.Insert(&ufs[i])
	}
	return ug
}

func groupFeature(ufs []URLFeature) *URLGroup {
	return nil
}

func URLFeatureSort(ufs []URLFeature) []URLFeature {

	sort.Slice(ufs, func(i, j int) bool {
		pc := compareStrs(ufs[i].PathFeature, ufs[j].PathFeature)
		if pc == -1 {
			return false
		}
		if pc == 1 {
			return true
		}

		if len(ufs[i].QueryFeature) > len(ufs[j].QueryFeature) {
			return false
		}

		if len(ufs[i].QueryFeature) < len(ufs[j].QueryFeature) {
			return true
		}

		for index := 0; index < len(ufs[i].QueryFeature); index++ {
			iqms := ufs[i].QueryFeature[index]
			jqms := ufs[j].QueryFeature[index]
			qkc := compareStr(iqms.K, jqms.K)
			if qkc == -1 {
				return false
			}
			if qkc == 1 {
				return true
			}
			qc := compareStrs(iqms.V, jqms.V)
			if qc == -1 {
				return false
			}
			if qc == 1 {
				return true
			}

		}
		return true
	})

	return ufs
}

// 0 等于
// -1 小于
// 1 大于
func compareStrs(as []string, bs []string) int {
	// 越长越小
	if len(as) < len(bs) {
		return -1
	}
	if len(as) > len(bs) {
		return 1
	}
	for i := 0; i < len(as); i++ {
		cv := compareStr(as[i], bs[i])
		if cv != 0 {
			return cv
		}
	}
	return 0
}

// 0 等于
// -1 小于
// 1 大于
func compareStr(a, b string) int {
	if len(a) > len(b) {
		return 1
	}
	if len(a) < len(b) {
		return -1
	}
	if a > b {
		return 1
	}
	if a < b {
		return -1
	}
	return 0
}

func DecomposeURLs(urls []string) []URLFeature {
	var ufs []URLFeature
	for _, u := range urls {
		uf, err := decomposeSingleURL(u, true)
		if err != nil {
			continue
		}
		ufs = append(ufs, *uf)
	}
	return ufs
}

func decomposeSingleURL(u string, exist bool) (*URLFeature, error) {
	pu, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	uf := URLFeature{
		URLRaw:       u,
		PathFeature:  []string{},
		QueryFeature: []QV{},
		Exist:        exist,
	}
	for _, s := range strings.Split(pu.Path, "/") {
		s = strings.ReplaceAll(s, ".html", "")
		if s != "" {
			uf.PathFeature = append(uf.PathFeature, s)
		}
	}
	queries := pu.Query()
	for _, q := range strings.Split(pu.RawQuery, "&") {
		if strings.Contains(q, "=") {
			qname := strings.Split(q, "=")[0]
			if qValue, ok := queries[qname]; ok {
				uf.QueryFeature = append(uf.QueryFeature, QV{K: qname, V: qValue})
			}
		}
	}
	return &uf, nil
}

// 这个方法纯属测试方便用的
func IsSub(u string, u2 string) bool {
	uf1, err := decomposeSingleURL(u, false)
	if err != nil {
		return false
	}
	uf2, err := decomposeSingleURL(u2, false)
	if err != nil {
		return false
	}
	return isSubFeature(*uf1, *uf2)
}

func isSubFeature(uf1 URLFeature, uf2 URLFeature) bool {
	if len(uf1.PathFeature) > len(uf2.PathFeature) {
		return false
	}
	for i, s := range uf1.PathFeature {
		if strings.ReplaceAll(s, ".html", "") != strings.ReplaceAll(uf2.PathFeature[i], ".html", "") {
			return false
		}
	}
	if len(uf1.PathFeature) < len(uf2.PathFeature) {
		return true
	}

	if len(uf1.QueryFeature) > len(uf2.QueryFeature) {
		return false
	}
	for i, s := range uf1.QueryFeature {
		if s.K != uf2.QueryFeature[i].K {
			return false
		} else {
			if len(s.V) != len(uf2.QueryFeature[i].V) {
				return false
			}
			for vi, vs := range s.V {
				if vs != uf2.QueryFeature[i].V[vi] {
					return false
				}
			}
		}
	}

	return true
}
