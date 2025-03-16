//go:build synctest

package try

import (
	"testing"
	"time"
)

func Test2(t *testing.T) {
	println("123")
	synctest.Run(func() {
		before := time.Now()
		time.Sleep(time.Second)
		after := time.Now()
		if d := after.Sub(before); d != time.Second {
			t.Fatalf("took %v", d)
		}
		t.Log("done")
	})
}
