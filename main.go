package main

import (
	"flag"
	"fmt"
	u "net/url"
	"os"
	"sync"

	"github.com/fatih/color"
)

// TestedURL - Represents a struct for collection data
type TestedURL struct {
	URL        u.URL
	Status     int
	LinkedUrls []u.URL
}

var wg sync.WaitGroup

var crawlResults = make(chan TestedURL)
var seedURLStatic u.URL

func main() {

	// Should show in colors?
	cliColors := flag.Bool("color", false, "Color output - true|false")

	// Get website
	cliWebsite := flag.String("url", "", "Website url")

	// Parse flags
	flag.Parse()

	// Parse and check if website it valid
	seedURL, err := u.ParseRequestURI(*cliWebsite)
	if err != nil {
		fmt.Println("Invalid website.")
		os.Exit(1)
	}
	seedURLStatic = *seedURL

	// Start the checker
	fmt.Printf("Getting links from: %s\n", seedURL)

	// Crawl site
	wg.Add(1)
	go Crawl(*seedURL)

	// Each results should output to the terminal
	go func() {
		for link := range crawlResults {

			status := "OK"
			if link.Status != 200 {
				status = "BAD"
			}

			statusMessage := status
			if *cliColors {
				statusMessage = color.GreenString(status)
				if status == "BAD" {
					statusMessage = color.RedString(status)
				}
			}

			fmt.Printf("%v - %v\n", statusMessage, link.URL.String())

		}
	}()

	// Close crawl results
	defer close(crawlResults)

	// Waitgroup wait
	wg.Wait()

}
