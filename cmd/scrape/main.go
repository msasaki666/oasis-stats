package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/msasaki666/oasis-stats/models"
	"github.com/otiai10/gosseract"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

func scrapeOasisUsageStats(db *gorm.DB) {
	if !inBusiness(int(time.Now().Weekday())) {
		log.Println("not in business time")
		os.Exit(0)
	}
	waitUntilRequiredTime(db)

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

func waitUntilRequiredTime(db *gorm.DB) {
	var lastStat models.UsageStat

	if tx := db.Order("scraped_at desc").First(&lastStat); tx.Error != nil {
		log.Fatal(tx.Error)
	}

	nextScrapingAt := lastStat.ScrapedAt.Add(15 * time.Minute)
	now := time.Now()
	if nextScrapingAt.Equal(now) || nextScrapingAt.After(now) {
		return
	} else {
		d := nextScrapingAt.Sub(now)
		if d > 10*time.Minute {
			log.Printf("wait limit is 10 minutes")
			os.Exit(0)
		}

		log.Printf("wait until %s\n", nextScrapingAt)
		time.Sleep(d)
		return
	}
}

func inBusiness(dayOfWeek int) bool {
	b, exist := os.LookupEnv("BUSINESS_TIME_PATTERN_" + strings.ToUpper(strconv.Itoa(dayOfWeek)))
	if !exist {
		return true
	}

	businessTimes := strings.Split(b, ",")
	now := time.Now()
	y, m, d := now.Date()
	start, end := fmt.Sprintf("%d-%s-%dT%s", y, m, d, businessTimes[0]), fmt.Sprintf("%d-%s-%dT%s", y, m, d, businessTimes[1])

	if now.After(parseTime(start)) && now.Before(parseTime(end)) {
		return true
	}

	return false
}

func parseTime(t string) time.Time {
	tt, err := time.Parse(time.RFC3339, t)
	if err != nil {
		log.Fatal(errors.WithStack(err).Error())
	}
	return tt
}
