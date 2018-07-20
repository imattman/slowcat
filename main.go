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

func main() {
	var (
		minDelay time.Duration
		maxDelay time.Duration
	)
	flag.DurationVar(&minDelay, "min", 1*time.Millisecond, "Minium delay between item output")
	flag.DurationVar(&maxDelay, "max", 10*time.Millisecond, "Maximum delay between item output")
	flag.Parse()

	data := make(chan []byte)

	go func() {
		defer close(data)

		if len(flag.Args()) < 1 {
			readRunes(os.Stdin, data)
			return
		}

		for _, fname := range flag.Args() {
			f, err := os.Open(fname)
			if err != nil {
				fatal("Error opening %s : %v", fname, err)
			}
			readRunes(f, data)
			f.Close()
		}
	}()

	delay := newDelayFunc(minDelay, maxDelay)
	err := writeSlow(os.Stdout, data, delay)
	if err != nil {
		fatal("Error writing %v", err)
	}
}

func fatal(msg string, args ...interface{}) {
	if !strings.HasSuffix(msg, "\n") {
		msg = msg + "\n"
	}
	fmt.Fprintf(os.Stderr, msg, args)
}

func readRunes(r io.Reader, data chan<- []byte) {
	scan := bufio.NewScanner(r)
	scan.Split(bufio.ScanRunes)
	for scan.Scan() {
		r := scan.Text()
		data <- []byte(r)
	}
}

func writeSlow(w io.Writer, data chan []byte, delayFunc func()) error {
	for b := range data {
		delayFunc()
		_, err := w.Write(b)
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
