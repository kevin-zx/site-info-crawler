package sitethrougher

import (
	"strings"
)

func GetPublicSuffix(ts []string) string {
	suffixMap := make(map[string]int)
	for _, t := range ts {
		for _, s := range ts {
			if len(t) > 0 && len(s) > 0 {
				suffix := getSuffix(s, t)
				if suffix == "" {
					break
				}
				//if suffix=="" {
				//	fmt.Println(t)
				//}
				if _, ok := suffixMap[suffix]; !ok {
					suffixMap[suffix] = 1
				} else {
					suffixMap[suffix] += 1
				}
			}
		}
	}
	publicSuffix := ""
	max := 0
	for s, i := range suffixMap {
		if i > max {
			max = i
			publicSuffix = s
		}
	}
	return publicSuffix
}

func getSuffix(t1 string, t2 string) string {
	t1p := strings.Split(t1, "")
	t2p := strings.Split(t2, "")
	c := len(t1p)
	suffix := ""
	for i := len(t2p)-1; i >= 0 && c > 0; i-- {
		c--
		if t2p[i] != t1p[c] {
			return suffix
		}
		for _, s := range splitText {
			if t2p[i] == s {
				return suffix
			}
		}
		suffix = t2p[i] + suffix
	}
	return suffix
}

