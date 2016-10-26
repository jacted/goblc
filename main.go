package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	u "net/url"
	"os"
	"sync"
)

type TestedUrl struct {
	Url        u.URL
	Status     int
	LinkedUrls []u.URL
}

var wg sync.WaitGroup

var crawlResults = make(chan TestedUrl)
var seedUrlStatic u.URL

func main() {

	// Should show in colors?
	cliColors := flag.Bool("color", false, "Color output - true|false")

	// Get website
	cliWebsite := flag.String("url", "", "Website url")

	// Parse flags
	flag.Parse()

	// Parse and check if website it valid
	seedUrl, err := u.ParseRequestURI(*cliWebsite)
	if err != nil {
		fmt.Println("Invalid website.")
		os.Exit(1)
	}
	seedUrlStatic = *seedUrl

	// Start the checker
	fmt.Printf("Getting links from: %s\n", seedUrl)

	// Crawl site
	wg.Add(1)
	go Crawl(*seedUrl)

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

			fmt.Printf("%v - %v\n", statusMessage, link.Url.String())

		}
	}()

	// Close crawl results
	defer close(crawlResults)

	// Waitgroup wait
	wg.Wait()

}
