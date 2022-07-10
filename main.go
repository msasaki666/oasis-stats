package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/otiai10/gosseract"
	"github.com/pkg/errors"
)

func scrapeOasisUsageStats() {
	client := gosseract.NewClient()
	defer client.Close()

	// Request the HTML page.
	res, err := http.Get("https://www.sportsoasis.co.jp/sh07/usage_stats/")
	if err != nil {
		log.Fatal(errors.WithStack(err).Error())
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(errors.WithStack(err).Error())
	}

	// Find the review items
	doc.Find("#point2 ul").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		url, exists := s.Find("li div img").First().Attr("src")
		if !exists {
			log.Fatal("利用状況を表示する要素が見つからん")
		}
		res, err := http.Get(url)
		if err != nil {
			log.Fatal(errors.WithStack(err).Error())
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		}
		b, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(errors.WithStack(err).Error())
		}
		if err = client.SetImageFromBytes(b); err != nil {
			log.Fatal(errors.WithStack(err).Error())
		}
		t, err := client.Text()
		if err != nil {
			log.Fatal(errors.WithStack(err).Error())
		}
		fmt.Printf("%s\n", t)
	})
}

func main() {
	scrapeOasisUsageStats()
}
