package main

import (
	"fmt"
	"net/http"
	u "net/url"
	str "strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var crawled = make(map[u.URL]bool)
var mutex = &sync.Mutex{}

var netClient = &http.Client{
	Timeout: time.Second * 20,
}

// Crawl - Function that crawls a given URL
func Crawl(url u.URL) {

	if isCrawled(url) != true {

		mutex.Lock()
		crawled[url] = true
		mutex.Unlock()

		testedChan, errChan := findLinks(url)

		var tested TestedURL
		var testedErr = false
		select {
		case t := <-testedChan:
			tested = t
			testedErr = false
		case err := <-errChan:
			fmt.Println(err.Error())
			testedErr = true
		}

		if !testedErr {
			crawlResults <- tested

			for _, child := range tested.LinkedUrls {
				if shouldCrawl(child) {
					wg.Add(1)
					go Crawl(child)
				}
			}
		}

	}

	defer wg.Done()

}

// findLinks - Function that finds every link on the page
func findLinks(url u.URL) (chan TestedURL, chan error) {
	c := make(chan TestedURL)
	errChan := make(chan error)

	go func() {

		req, err := http.NewRequest("GET", url.String(), nil)
		req.Header.Add("Accept-Encoding", "identity")
		req.Close = true

		res, err := netClient.Do(req)
		if err != nil {
			errChan <- err
			return
		}
		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			errChan <- fmt.Errorf("Error parsing %s", url.String())
			return
		}

		linkedUrls := make([]u.URL, 0)
		doc.Find("a").Each(func(_ int, sel *goquery.Selection) {
			href, _ := sel.Attr("href")
			parsed, err := u.Parse(stripLinks(href))
			if err != nil {
				errChan <- err
				return
			}
			linkedUrls = append(linkedUrls, *url.ResolveReference(parsed))
		})

		c <- TestedURL{url, res.StatusCode, linkedUrls}
		close(c)

	}()

	return c, errChan
}

func stripLinks(s string) string {
	lastIndex := str.LastIndex(s, "#")
	if lastIndex != -1 {
		s = s[0:lastIndex]
	}
	s = str.TrimSuffix(s, "/")
	return s
}

func isCrawled(url u.URL) bool {
	mutex.Lock()
	_, alreadyCrawled := crawled[url]
	mutex.Unlock()
	return alreadyCrawled
}

func shouldCrawl(potential u.URL) bool {
	return seedURLStatic.Host == potential.Host
}
