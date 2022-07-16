package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/msasaki666/oasis-stats/models"
	"github.com/otiai10/gosseract"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func scrapeOasisUsageStats(db *gorm.DB) {
	client := gosseract.NewClient()
	defer client.Close()

	statPattern := regexp.MustCompile(`\d+`)

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
	var stats []int
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
		stat := statPattern.FindString(t)
		statInt, err := strconv.Atoi(stat)
		if err != nil {
			log.Fatal(errors.WithStack(err).Error())
		}
		stats = append(stats, statInt)
	})
	m := models.UsageStat{Female: stats[0], Male: stats[1], ScrapedAt: time.Now()}
	db.Create(&m)
}

func main() {
	dsn, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("set DATABASE_URL")
	}
	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{},
	)
	if err != nil {
		log.Fatal(errors.WithStack(err).Error())
	}
	if err = db.AutoMigrate(models.MigrationTargets()...); err != nil {
		log.Fatal(errors.WithStack(err).Error())
	}
	scrapeOasisUsageStats(db)
}
