package try

import (
	"fmt"
	"testing"
	"testing/synctest"
	"time"
)

func Test2(t *testing.T) {
	fmt.Println("123")
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
