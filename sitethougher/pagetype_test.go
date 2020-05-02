package sitethougher

import "testing"

func Test_breadUrlJudge(t *testing.T) {
	urls := []string{
		//"http://www.jsszsf.cn/list-9-1.html",
		"http://www.yiheshoot.com/lianxiwomen.html",
		"http://www.xyfsjg.com/gongsijieshao","http://www.lightjun.com/index.php/list/23","http://www.jhccq.com/a/contact/",
		"http://m.whgcmc.cn/p125.html","http://www.whtdld.com/a/chanpin/hubeijiaotougaosulujiance/",
		"http://www.zhouheiya.net/html/lianxiwomen/lianxifangshi/","http://www.scxzlhjpx.cn/yonghuliuyan",
		"http://www.chenghaomaoyi.com/list.asp?classid=16",
	}
	for _, url := range urls {
		if got := breadUrlJudge(url); got != true {
			t.Errorf("breadUrlJudge() = %s, want %v", url, true)
		}
	}

	furls := []string{"http://www.cqweixing.cn/page7","http://www.cqsakj.com/page5?product_category=4&brd=1"}
	for _, url := range furls {
		if got := breadUrlJudge(url); got != false {
			t.Errorf("breadUrlJudge() = %s, want %v", url, false)
		}
	}
}

func Test_judgeHome(t *testing.T) {
	type args struct {
		absURL string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1",
			args: args{"http://www.tpsyyq.com"},
			want: true,
		},
		{
			name: "2",
			args: args{"http://www.tpsyyq.com/"},
			want: true,
		},
		{
			name: "2",
			args: args{"http://www.tpsyyq.com/t/19.html"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := judgeHome(tt.args.absURL); got != tt.want {
				t.Errorf("judgeHome() = %v, want %v", got, tt.want)
			}
		})
	}
}