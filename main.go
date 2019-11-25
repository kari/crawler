package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gocolly/colly"
)

// Client wraps a http.Client
type Client struct {
	httpClient *http.Client
	crawled    map[string]bool
}

// NewClient creates an instance of Client
func NewClient() *Client {
	c := &Client{
		httpClient: &http.Client{
			CheckRedirect: ObserveRedirects,
		},
		crawled: make(map[string]bool),
	}
	return c
}

// ObserveRedirects logs all redirects
// Most redirects should be just no-www and to https
func ObserveRedirects(req *http.Request, via []*http.Request) error {
	fmt.Printf("%s --> ", via[len(via)-1].URL)
	// to check the actual 3xx code, this should happen at Transport, https://stackoverflow.com/questions/24577494/how-to-get-the-http-redirect-status-codes-in-golang
	return nil
}

func main() {
	collector := colly.NewCollector(
		colly.AllowedDomains("kalifi.org"),
	)

	client := NewClient()

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		collector.Visit(e.Request.AbsoluteURL(link))
		client.fetch(e.Request.AbsoluteURL(link))
	})

	collector.Visit("https://kalifi.org/sitemap.html")
}

func (c Client) fetch(url string) {
	if _, ok := c.crawled[url]; ok {
		return
	}
	c.crawled[url] = true

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return
	}
	if req.URL.Host == "kalifi.org" {
		return
	}
	// resp throws an error for unsupported protocol scheme which
	// could be caught as well
	if !(req.URL.Scheme == "http" || req.URL.Scheme == "https") {
		return
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		switch {
		case strings.HasSuffix(err.Error(), "connection refused"):
			fmt.Printf("%s: connection refused\n", url)
			return
		case strings.HasSuffix(err.Error(), "no such host"):
			fmt.Printf("%s: no such host\n", url)
			return
		case strings.HasSuffix(err.Error(), "i/o timeout"):
			fmt.Printf("%s: i/o timeout\n", url)
			return
		}
		log.Println(err)
		return
	}
	fmt.Printf("%s (%d)\n", resp.Request.URL, resp.StatusCode)
}
