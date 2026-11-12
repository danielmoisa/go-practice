package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/yeka/zip"
)

var (
	zipfile     string
	dictionary  string
	concurrency int
)

func init() {
	flag.StringVar(&zipfile, "f", "", "Open zipfile")
	flag.StringVar(&dictionary, "d", "", "Open pass dictionary")
	flag.IntVar(&concurrency, "c", runtime.NumCPU(), "Number of workers to use")
}

func main() {
	flag.Parse()

	if zipfile == "" || dictionary == "" {
		println("Please " + os.Args[0] + " -h")
		os.Exit(0)
	}

	word := make(chan string)
	found := make(chan string)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go zipCrackWorker(word, found, &wg)
	}

	dictfile, err := os.Open(dictionary)
	if err != nil {
		log.Fatalln(err)
	}

	defer dictfile.Close()

	scanner := bufio.NewScanner(dictfile)
	if err := scanner.Err(); err != nil {
		log.Fatalf("reading dictionary: %v", err)
	}

	go func() {
		for scanner.Scan() {
			pass := strings.TrimSpace(scanner.Text())
			if pass == "" {
				continue
			}
			word <- pass
		}
		close(word)
	}()

	done := make(chan bool)

	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case f := <-found:
		println("[+] Found password")
		println("[+] Password =", f)

		return
	case <-done:
		println("[+] Password not found")

		return
	}
}

func zipCrackWorker(word <-chan string, found chan<- string, wg *sync.WaitGroup) {
	ziper, err := zip.OpenReader(zipfile)
	if err != nil {
		log.Fatal(err)
	}
	defer ziper.Close()
	defer wg.Done()

	for w := range word {
		correct := true
		for _, z := range ziper.File {
			z.SetPassword(w)
			rc, err := z.Open()
			if err != nil {
				correct = false
				break
			}

			_, err = io.Copy(io.Discard, rc) // forces CRC-32 check
			rc.Close()
			if err != nil {
				correct = false
				break
			}
		}

		if correct {
			select {
			case found <- w:
			default:
			}
			return
		}
	}
}
