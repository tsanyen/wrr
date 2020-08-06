package wrr_test

import (
	"github.com/tsanyen/wrr"
	"testing"
)

func TestPut(t *testing.T) {
	r := wrr.New(wrr.WithCap(0))

	if _, err := r.Round(); err != wrr.ErrEmptyNodes {
		t.Errorf("Round() with empty nodes got success; want err")
	}

	r.Put("a", 3)
	r.Put("b", 1)
	r.Put("c", 1)

	z := ""
	for i := 0; i < 5; i++ {
		x, _ := r.Round()
		z += x.(string)
	}
	if z != "abaca" {
		t.Errorf("5 times Round() gots %s; want abaca", z)
	}

	r.Put("d", 3)
	for i := 0; i < 8; i++ {
		r.Round()
	}

	r.Remove("d")
	for i := 0; i < 7; i++ {
		r.Round()
	}

	r.Remove("a")

	r.Gets(0)

	r.Put("a", 2)

	z = ""
	for i := 0; i < 4; i++ {
		x, _ := r.Round()
		z += x.(string)
	}
	if z != "abca" {
		t.Errorf("4 times Round() gots %s; want abca", z)
	}

}
