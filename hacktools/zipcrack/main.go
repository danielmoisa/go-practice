package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/yeka/zip"
)

func main() {
	// Set the cmd flags
	zipfile := flag.String("f", "", "zip file to open")
	dictionary := flag.String("d", "", "password dictionary")
	flag.Parse()

	// Check for flags
	if *zipfile == "" || *dictionary == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Open the dictionary file.
	f, err := os.Open(*dictionary)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// CHANNELS: goroutines communicate by sending values through channels.
	// words: unbuffered, so a send blocks until a worker is ready to receive.
	//        This means the dictionary reader only reads as fast as workers consume.
	words := make(chan string)

	// result: buffered with room for 1 value. A worker can drop its answer here
	//         without waiting for main to be ready to read it.
	result := make(chan string, 1) // buffered so a late winner never blocks

	// stop: never receives a value. We only ever close it, and a closed channel
	//       is "readable" immediately, so a select on it fires as a broadcast signal.
	stop := make(chan struct{})

	// sync.Once guarantees the function inside Do runs only once, even if
	// several workers succeed at the same time. Here it protects close(stop),
	// because closing an already-closed channel causes a panic.
	var once sync.Once

	// WaitGroup counts running goroutines. wg.Wait() blocks until all of them finish.
	var wg sync.WaitGroup

	// Start one worker goroutine per CPU core.
	// wg.Go starts the function in a goroutine and tracks it in the WaitGroup.
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Go(func() {
			// Each worker opens its own copy of the zip. SetPassword changes the
			// state of zip.File objects, so sharing them between workers would cause races.
			zr, err := zip.OpenReader(*zipfile)
			if err != nil {
				log.Fatal(err)
			}
			defer zr.Close()

			// range over a channel receives values until the channel is closed.
			// Each worker takes the next available word.
			for w := range words {
				if checkPassword(zr.File, w) {
					// select with default is a non-blocking send: if the buffer is
					// full, skip it instead of waiting.
					select {
					case result <- w:
					default:
					}

					// Tell the dictionary reader to stop sending new words.
					once.Do(func() { close(stop) })

					// This worker is finished, so it returns.
					return
				}
			}
		})
	}

	// Feeder goroutine: reads the dictionary and sends each word to the workers.
	go func() {
		// Closing words when this function ends tells workers there are no more
		// words, which ends their range loops.
		defer close(words)

		scanner := bufio.NewScanner(f)
		for scanner.Scan() { // reads one line at a time
			pass := strings.TrimSpace(scanner.Text()) // remove spaces and newline characters
			if pass == "" {
				continue
			}

			// select waits on several channel operations and runs whichever is ready first.
			select {
			case words <- pass: // send the word to a worker
			case <-stop: // or stop early if a password has already been found
				return
			}
		}

		// scanner.Scan returns false at end of file or on error. Check which one it was.
		if err := scanner.Err(); err != nil {
			log.Fatalf("reading dictionary: %v", err)
		}
	}()

	// Block until every worker has returned. This happens when either the
	// password was found, or the dictionary ran out and words was closed.
	wg.Wait()

	// After all workers finish, check whether a result was stored.
	// select with default does not block: if nothing is in result, we take the default branch.
	select {
	case pw := <-result:
		fmt.Println("[+] Password found")
		fmt.Println("[+] Password =", pw)
	default:
		fmt.Println("[-] Password not found")
	}
}

// checkPassword returns true if pass unlocks every file in the zip.
func checkPassword(files []*zip.File, pass string) bool {
	for _, f := range files {
		f.SetPassword(pass)

		// Open fails quickly for many wrong passwords.
		rc, err := f.Open()
		if err != nil {
			return false
		}

		// Reading the whole file makes the library check the CRC-32 checksum.
		// A wrong password usually produces garbage data, which fails the check.
		_, err = io.Copy(io.Discard, rc) // io.Discard throws the data away; we only care about errors
		rc.Close()
		if err != nil {
			return false
		}
	}
	return true
}
