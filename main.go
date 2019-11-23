package main

import (
	"fmt"
	"net/http"

	"github.com/gocolly/colly"
)

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("kalifi.org"),
	)
	c2 := colly.NewCollector(
		colly.DisallowedDomains("kalifi.org"),
	)

	c2.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})

	c.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting", req.URL)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		c.Visit(e.Request.AbsoluteURL(link))
		c2.Visit(e.Request.AbsoluteURL(link))
	})

	c2.OnRequest(func(r *colly.Request) {
		r.Ctx.Put("url", r.URL.String())
	})

	c2.OnResponse(func(resp *colly.Response) {
		fmt.Printf("%s -> %s (%d)\n", resp.Ctx.Get("url"), resp.Request.URL, resp.StatusCode)

	})

	c2.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r.StatusCode, "\nError:", err)
	})

	c.Visit("https://kalifi.org/sitemap.html")
}
