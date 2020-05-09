package tokenbucket

import (
	"testing"
	"time"
)

func TestNewBucket(t *testing.T) {

	b := NewBucket(1, 2*time.Second)
	if b.Take(1) == 0 {
		t.Log("1 ok")
	}
	if b.Take(1) == 0 {
		t.Log("2 ok")
	}
	if b.Take(1) == 0 {
		t.Log("3 ok")
	}
	time.Sleep(2 * time.Second)

	if b.Take(1) == 0 {
		t.Error("4 sleep but not not ok")
	} else {
		t.Log("4 when sleep, ok")
	}
	if b.Take(1) == 0 {
		t.Log("5 ok")
	}
}
