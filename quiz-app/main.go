package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	// Add the csv flag
	csvFileName := flag.String("csv", "problems.csv", "a csv file in the format of question and answer")
	limit := flag.Int("limit", 10, "time limit for the quiz in seconds")
	flag.Parse()

	// Read the csv file
	csvFile, err := os.Open(*csvFileName)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %s\n", *csvFileName)
	}
	defer csvFile.Close()

	// Parse the csv
	r := csv.NewReader(csvFile)
	lines, err := r.ReadAll()
	if err != nil {
		log.Fatal("Error reading CSV blocks", err)
	}

	// Init timer
	timer := time.NewTimer(time.Duration(*limit) * time.Second)

	// Run the quiz
	correct := 0
	for i, line := range lines {
		fmt.Printf("Problem #%d: %s = ", i+1, line[0])

		// Create a channel for timer
		answerCh := make(chan string)

		// Launch the goroutine
		go func() {
			var answer string
			fmt.Scanf("%s\n", &answer)

			answerCh <- answer
		}()

		// Select the timer if ends, else continue with answer
		select {
		case <-timer.C:
			fmt.Printf("\nTime's up! You scored %d/%d.\n", correct, len(lines))
			return
		case answer := <-answerCh:
			if strings.TrimSpace(answer) == strings.TrimSpace(line[1]) {
				correct++
			}
		}

	}

	fmt.Printf("You scored %d out of %d.\n", correct, len(lines))
}
