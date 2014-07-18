package main

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"strconv"
	"strings"
	_ "sync"
	"time"
)

var (
	base, nums int = 10, 10
	output     []string
)

var timeout = errors.New("timeout")

func none01(ch rune) bool {
	switch ch {
	case '0', '1':
		return false
	default:
		return true
	}
}

func isZeroOne(v *big.Int) bool {
	return -1 == strings.IndexFunc(v.String(), none01)
}

/* naive implementation */
func zero1(N int) (string, error) {
	v := big.NewInt(int64(N))

	for {
		if output[N-base] != "" {
			return "", timeout
		}
		if isZeroOne(v) {
			return v.String(), nil
		}
		v.Add(v, big.NewInt(int64(N)))
	}

	return v.String(), nil
}

func nextNineOnes(v *big.Int, N int) *big.Int {
	if len(fmt.Sprint(N)) != len(fmt.Sprint(N+1)) {
		sv := strings.Repeat("111111111", len(fmt.Sprint(N)))
		v.SetString(sv, 10)
		return v
	}

	if v.Cmp(big.NewInt(111111111)) < 0 {
		v.SetInt64(111111111)
		return v
	}

	for {
		sv := v.String()
		vbin := new(big.Int)
		vbin.SetString(sv, 2)
		vbin.Add(vbin, big.NewInt(1))
		sv = fmt.Sprintf("%b", vbin)
		v.SetString(sv, 10)

		c1 := strings.Count(sv, "1")
		if c1%9 == 0 {
			return v
		}
		last0s := len(sv) - (strings.LastIndex(sv, "1") + 1)
		need1s := 9 - (c1 % 9)
		if last0s >= need1s {
			sv = sv[:(len(sv)-need1s)] + strings.Repeat("1", need1s)
			v.SetString(sv, 10)
			return v
		}
	}
}

func next01Number(N int) string {
	sv := fmt.Sprint(N)
	idx := strings.IndexFunc(sv, none01)

	if idx >= 0 {
		vbin, err := strconv.ParseInt("0"+sv[:idx], 2, 0)
		if err != nil {
			fmt.Println("not ok", sv, idx, err)
			return "-1"
		}
		vbin++
		sv = fmt.Sprintf("%b", vbin)
		if len(sv) > idx {
			sv = sv + strings.Repeat("0", len(sv)-idx)
		}
	}

	return sv
}

func reduce25(N int) (int, int) {
	non0 := func(ch rune) bool {
		switch ch {
		case '0':
			return false
		default:
			return true
		}
	}

	sN := fmt.Sprintf("%d", N)
	last0s := len(sN) - (strings.LastIndexFunc(sN, non0) + 1)
	if last0s > 0 {
		n1, _ := strconv.ParseInt(sN[:len(sN)-last0s], 10, 0)
		N = int(n1)
	}

	sN2 := fmt.Sprintf("%b", N)
	last0s2 := len(sN2) - (strings.LastIndexFunc(sN2, non0) + 1)
	if last0s2 > 0 {
		n1, _ := strconv.ParseInt(sN2[:len(sN2)-last0s2], 2, 0)
		N = int(n1)
	}
	by5s := 0
	for N%5 == 0 {
		N /= 5
		by5s++
	}

	if last0s2 > by5s {
		last0s += last0s2
	} else {
		last0s += by5s
	}

	return N, last0s
}

func zero12(N int) (string, error) {
	savedN := N
	N, last0s := reduce25(N)

	// fmt.Printf("in testing updated N=%d, last0s=%d\n", N, last0s)

	upper := new(big.Int)
	upper.SetString("999999999999999999999999999999999999999", 10)

	v := new(big.Int)
	v.SetString(next01Number(N), 10)
	sv := v.String()
	vbin := new(big.Int)
	vbin.SetString(sv, 2)

	// fmt.Println("in testing start with:\t", N, v, last0s)
	for {
		if output[savedN-base] != "" {
			return "", timeout
		}

		rem := new(big.Int)
		if rem.Rem(v, big.NewInt(int64(N))); rem.Cmp(big.NewInt(0)) == 0 {
			// fmt.Println("rem res:\t", v, sv, N, rem)
			if last0s > 0 {
				sv += strings.Repeat("0", last0s)
			}
			return sv, nil
		} else if v.Cmp(upper) > 0 {
			return "", timeout
		}
		if N%9 == 0 {
			v = nextNineOnes(v, N)
			sv = v.String()
		} else {
			vbin.Add(vbin, big.NewInt(1))
			sv = fmt.Sprintf("%b", vbin)
			v.SetString(sv, 10)
		}
	}

	return v.String(), nil
}

func zero120(N int) (string, error) {
	savedN := N
	N, last0s := reduce25(N)

	sv := next01Number(N)
	v, err := strconv.ParseUint(sv, 10, 0)
	v2, err2 := strconv.ParseUint(sv, 2, 0)
	if err != nil || err2 != nil {
		return "", errors.New("strconv.ParseUint failure")
	}

	for {
		if v%uint64(N) == 0 {
			break
		}

		v2++
		if v2 > 524287 {
			return "", fmt.Errorf("zero120 overflowing for %d", savedN)
		}
		// if v2 > 524287 {
		// in binary 1111111111111111111 same length as math.MaxUint64
		// 1<<64 - 1 == 9223372036854775808L width 19
		// '1' * 19  == 1111111111111111111b = 524287 in decimal
		// return "-1"
		// }

		sv = fmt.Sprintf("%b", v2)
		v, err = strconv.ParseUint(sv, 10, 64)
		if err != nil {
			fmt.Println("overflowing and give up:", err)
			return "", err
		}
	}

	sv = fmt.Sprint(v)
	if last0s > 0 {
		sv += strings.Repeat("0", last0s)
	}
	return sv, nil
}

type worker func(int) (string, error)
type pair01 struct {
	N   int
	res string
	ti  time.Duration
	sig string
}

func (info pair01) String() string {
	line := fmt.Sprintf("%d:\t%s", info.N, info.res)
	if info.ti > 0 {
		line += "\t" + info.ti.String()
	}
	if info.sig != "" {
		line += " by " + info.sig
	}
	return line
}

func handle(N int, func1 worker, sig string,
	outch chan<- pair01, done func()) {
	if done != nil {
		defer done()
	}

	// fmt.Println("computing for:", N)
	t1 := time.Now()
	res, err := func1(N)
	if err != nil {
		if err != timeout {
			fmt.Println(err)
		}
		return
	}
	t2 := time.Now()
	ti := t2.Sub(t1)

	line := pair01{N: N, res: res, sig: sig}
	if ti > 100*time.Millisecond {
		line.ti = ti
	}

	outch <- line
}

func scheduler(outch chan<- pair01) {
	// waiter := &sync.WaitGroup{}
	done := make(chan struct{}, 20)
	fmt.Println("start with ", runtime.NumGoroutine(), "goroutines",
		cap(done), len(done))

	enqueued := 0
	finished := 0
	N := base
	for ; N <= base+nums; N++ {
		for enqueued-finished >= cap(done) {
			// fmt.Println("running:", running, cap(done), len(done))
			<-done
			finished++
		}

		finish := func() { done <- struct{}{} }

		for sig, func1 := range map[string]worker{
			"zero12":  zero12,
			"zero120": zero120,
			"zero1":   zero1,
		} {
			go handle(N, func1, sig, outch, finish)
			enqueued++
		}

		// fmt.Println("enqueued", N, "into goroutine", runtime.NumGoroutine(), "running,")
	}

	fmt.Printf("all %d numbers scheduled, waiting all finished (%d so far)... %d goroutines\n",
		enqueued, finished, runtime.NumGoroutine())

	for ; finished < enqueued; finished++ {
		<-done
	}

	close(outch)
}

func waitResults(outch <-chan pair01) {
	nextwaiting := base
	waitingnumbers := make([]string, 0, 10)
	output = make([]string, nums+1)
	const layout = "2006 01 02-150405.000 周一:"

	finishednumbers := 0
	lasttime := time.Now()
	lastwaitprint := 0
WAITALL:
	for {
		select {
		case line, ok := <-outch:
			if !ok {
				break WAITALL
			}
			if output[line.N-base] != "" {
				saved := strings.Fields(output[line.N-base])[0]
				if line.res != saved {
					fmt.Println("different result for:\t", line.N, saved, line.res, "by", line.sig)
				}
			} else {
				output[line.N-base] = line.res
				finishednumbers++

				if line.ti > 0 {
					waitingnumbers = append(waitingnumbers, line.String())
					fmt.Println(time.Now().Format(layout), "got", line,
						"nextwaiting", nextwaiting, ",",
						runtime.NumGoroutine(), "goroutines running")
				}

				if finishednumbers == len(output) {
					// all numbers finished, give up others slow computing...
					fmt.Println("completed.")
					break WAITALL
				}
			}

			if line.N == nextwaiting {
				for ; nextwaiting <= base+nums; nextwaiting++ {
					if output[nextwaiting-base] == "" {
						break
					}
				}
			}

		case <-time.After(1 * time.Second):
			t1 := time.Now()
			if t1.Sub(lasttime) > 12*time.Second || nextwaiting != lastwaitprint {
				fmt.Println(time.Now().Format(layout), "waiting on", nextwaiting,
					"finished:", finishednumbers, "of", len(output), ",",
					runtime.NumGoroutine(), "goroutines running")

				lasttime = t1
				lastwaitprint = nextwaiting
			}
		}
	}

	if len(waitingnumbers) > 0 {
		fmt.Println("\nthe slowest numbers:\n")
		for _, wline := range waitingnumbers {
			fmt.Println(wline)
		}
	}

	if nums <= 100 {
		fmt.Println("\ntesting numbers:")
		for i, line := range output {
			fmt.Printf("%d =>\t%s\n", base+i, line)
		}
	}
}

func main() {
	fmt.Println("start with ", runtime.NumGoroutine(), "goroutines")

	/*
		for _, N := range []int{4, 5, 6, 7, 8, 10, 11, 12, 19, 25, 50, 100} {
			fmt.Println("testing zero12:\t", N, zero12(N))
			fmt.Println("testing zero120:\t", N, zero120(N))
			fmt.Println("testing zero1:\t", N, zero1(N))
		} */

	runtime.GOMAXPROCS(runtime.NumCPU())

	base, _ = strconv.Atoi(os.Args[1])
	if len(os.Args) >= 3 {
		nums, _ = strconv.Atoi(os.Args[2])
	}

	outch := make(chan pair01, 4*10)
	go scheduler(outch)
	waitResults(outch)
}
