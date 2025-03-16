package try

import (
	"testing"
	"testing/synctest"
	"time"
)

func Test(t *testing.T) {
	synctest.Run(func() {
		before := time.Now()
		time.Sleep(time.Second)
		after := time.Now()
		if d := after.Sub(before); d != time.Second {
			t.Fatalf("took %v", d)
		}
	})
}
