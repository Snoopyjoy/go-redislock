package redislock

import (
	"os"
	"sync"
	"testing"
)

var lf ILockFactory

func TestMain(m *testing.M) {
	// var err error
	// lf, err = NewLockFactory()
	m.Run()
}

func Test_Lock(t *testing.T) {
	hostname, _ := os.Hostname()
	t.Log(hostname)
}

func Test_hashNum(t *testing.T) {
	ah := hashNum("a")
	if ah == 0 {
		t.Error("hash a equal zero")
	}
	ab := hashNum("b")
	if ab == 0 {
		t.Error("hash a equal zero")
	}
	ac := hashNum("c")
	if ac == 0 {
		t.Error("hash a equal zero")
	}
}

func Test_idGen(t *testing.T) {
	ids := [10]string{}
	wg := &sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(_i int) {
			v := idGen()
			ids[_i] = v
			wg.Done()
		}(i)
	}
	wg.Wait()

	t.Log(ids)
	for i := 0; i < 9; i++ {
		v1 := ids[i]
		for j := i + 1; j < 10; j++ {
			v2 := ids[j]
			if v1 == v2 {
				t.Errorf("id conflict, %d:%v, %d:%v", i, v1, j, v2)
			}
		}
	}
}
