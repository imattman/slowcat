package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	splitBytes := flag.Bool("b", true, "No partitioning; read input as bytes")
	splitRunes := flag.Bool("c", false, "Partition input at UTF-8 character boundaries")
	splitWords := flag.Bool("w", false, "Partition input at word boundaries")
	splitLines := flag.Bool("l", false, "Partition input at line boundaries")
	minDelay := flag.Duration("min", 10*time.Millisecond, "Minimum delay between items output")
	maxDelay := flag.Duration("max", 100*time.Millisecond, "Maximum delay between items output")
	uniform := flag.Duration("d", 0*time.Millisecond, "Sets min and max to same value for uniform delay between items")
	flag.Parse()

	if *uniform > 0 {
		*minDelay = *uniform
		*maxDelay = *uniform
	}

	if *maxDelay < *minDelay {
		*maxDelay = *minDelay
	}

	ctx := context.Background()
	ticker := delayedTicker(ctx, *minDelay, *maxDelay)

	if len(flag.Args()) < 1 {
		scan := bufio.NewScanner(os.Stdin)
		split, sep := splitterFunc(*splitLines, *splitWords, *splitRunes, *splitBytes)
		scan.Split(split)
		copySlow(ctx, os.Stdout, scan, sep, ticker)
		return
	}

	for _, fname := range flag.Args() {
		f, err := os.Open(fname)
		if err != nil {
			fatal("Error opening %s : %v", fname, err)
		}

		scan := bufio.NewScanner(f)
		split, sep := splitterFunc(*splitLines, *splitWords, *splitRunes, *splitBytes)
		scan.Split(split)

		err = copySlow(ctx, os.Stdout, scan, sep, ticker)
		f.Close()
		if err != nil {
			fatal("Error writing %v", err)
		}
	}
}

func copySlow(ctx context.Context, w io.Writer, scan *bufio.Scanner, sep []byte, ticker <-chan struct{}) error {
	firstItem := true

	for scan.Scan() {
		bs := scan.Bytes()
		err := scan.Err()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// tick precedes write
		select {
		case <-ctx.Done():
			return nil
		case <-ticker: // keep going
		}

		_, err = w.Write(bs)
		firstItem = false
		if err != nil {
			return err
		}

		if !firstItem {
			_, err = w.Write(sep)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func delayedTicker(ctx context.Context, min, max time.Duration) <-chan struct{} {
	delay := newDelayFunc(min, max)
	tickCh := make(chan struct{})

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case tickCh <- struct{}{}:
				delay()
			}
		}
	}()

	return tickCh
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

func splitterFunc(splitLines, splitWords, splitRunes, splitBytes bool) (split bufio.SplitFunc, sep []byte) {
	switch {
	case splitLines:
		return bufio.ScanLines, []byte("\n")
	case splitWords:
		return bufio.ScanWords, []byte(" ")
	case splitRunes:
		return bufio.ScanRunes, nil
	default:
		return bufio.ScanBytes, nil
	}
}

func fatal(msg string, args ...interface{}) {
	if !strings.HasSuffix(msg, "\n") {
		msg = msg + "\n"
	}
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(1)
}
