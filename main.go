package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

const defaultDelay = 10 * time.Millisecond

func main() {
	var (
		minDelay time.Duration
		maxDelay time.Duration
	)
	flag.DurationVar(&minDelay, "d", defaultDelay, "Uniform delay between item output")
	flag.DurationVar(&minDelay, "min", defaultDelay, "Minimum variable delaybetween item output")
	flag.DurationVar(&maxDelay, "max", 0, "Maximum variable delay between item output (defaults to 'min')")
	flag.Parse()

	if maxDelay < minDelay {
		maxDelay = minDelay
	}

	delayFunc := newDelayFunc(minDelay, maxDelay)

	if len(flag.Args()) < 1 {
		copySlow(os.Stdout, os.Stdin, delayFunc)
		return
	}

	for _, fname := range flag.Args() {
		f, err := os.Open(fname)
		if err != nil {
			fatal("Error opening %s : %v", fname, err)
		}

		err = copySlow(os.Stdout, f, delayFunc)
		f.Close()
		if err != nil {
			fatal("Error writing %v", err)
		}
	}
}

func copySlow(w io.Writer, r io.Reader, delayFunc func()) error {
	buf := bufio.NewReader(r)
	bs := make([]byte, 1)

	for {
		b, err := buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		delayFunc()
		bs[0] = b
		_, err = w.Write(bs)
		if err != nil {
			return err
		}
	}

	return nil
}

func newDelayFunc(min time.Duration, max time.Duration) func() {
	if min == max {
		return func() {
			time.Sleep(min)
		}
	}

	rand.Seed(time.Now().UnixNano())
	shifted := int(max) - int(min)

	return func() {
		time.Sleep(time.Duration(rand.Intn(shifted) + int(min)))
	}
}

func fatal(msg string, args ...interface{}) {
	if !strings.HasSuffix(msg, "\n") {
		msg = msg + "\n"
	}
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(1)
}
