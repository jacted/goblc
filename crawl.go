package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	u "net/url"
	str "strings"
	"sync"
	"time"
)

type TestedUrl struct {
	Url        u.URL
	Status     int
	LinkedUrls []u.URL
}

var crawled map[u.URL]bool = make(map[u.URL]bool)

var mutex = &sync.Mutex{}

var netClient = &http.Client{
	Timeout: time.Second * 20,
}

func Crawl(url u.URL) {
	wg.Add(1)
	go worker(url)
}

func worker(url u.URL) {

	if isCrawled(url) != true {

		mutex.Lock()
		crawled[url] = true
		mutex.Unlock()

		testedChan, errChan := findLinks(url)

		var tested TestedUrl
		var testedErr bool = false
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
					go worker(child)
				}
			}
		}

	}

	defer wg.Done()

}

func findLinks(url u.URL) (chan TestedUrl, chan error) {
	c := make(chan TestedUrl)
	errChan := make(chan error)

	go func() {

		res, err := netClient.Get(url.String())
		if err != nil {
			errChan <- err
		}

		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			errChan <- err
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

		c <- TestedUrl{url, res.StatusCode, linkedUrls}
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

	if !alreadyCrawled {
		return false
	}
	return true
}

func shouldCrawl(potential u.URL) bool {
	return seedUrlStatic.Host == potential.Host
}
