package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gocolly/colly"
)

// Client wraps a http.Client
type Client struct {
	httpClient *http.Client
	crawled    map[string]bool
	host       string
}

// LogRedirects adds logging for http.Client redirects
type LogRedirects struct {
	Transport http.RoundTripper
}

// RoundTrip logs redirect status
func (l LogRedirects) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t := l.Transport
	if t == nil {
		t = http.DefaultTransport
	}
	resp, err = t.RoundTrip(req)
	if err != nil {
		return
	}
	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect:
		fmt.Printf("%s (%d) --> ", req.URL, resp.StatusCode)
	}
	return
}

// NewClient creates an instance of Client
func NewClient(host string) *Client {
	c := &Client{
		httpClient: &http.Client{
			Transport: LogRedirects{},
			// CheckRedirect: ObserveRedirects,
		},
		crawled: make(map[string]bool),
		host:    host,
	}
	return c
}

var path = flag.String("url", "https://kalifi.org/sitemap.html", "url from where to start crawling the site and check outbound links")

func main() {
	u, err := url.Parse(*path)
	if err != nil {
		log.Fatal(err)
	}

	client := NewClient(u.Hostname())

	collector := colly.NewCollector(
		colly.AllowedDomains(u.Hostname()),
	)

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		collector.Visit(e.Request.AbsoluteURL(link))
		client.fetch(e.Request.AbsoluteURL(link))
	})

	collector.Visit(u.String())
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

	// only check outbound links
	if req.URL.Host == c.host {
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
