package gothoughtsite

import (
	"fmt"
	"testing"
	"time"
)

func TestRunWithParams(t *testing.T) {

	gotLinkMap, err := RunWithParams("http://www.daqing886.com/", 100,
		time.Second*1, 1)
	if err != nil {
		panic(err)
	}

	for _, l := range gotLinkMap {
		fmt.Printf("%s , %d\n", l.AbsURL, l.QuoteCount)
	}
}
