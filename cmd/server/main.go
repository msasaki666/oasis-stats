package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/msasaki666/oasis-stats/models"
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
	var stats []models.UsageStat
	if tx := db.Find(&stats); tx.Error != nil {
		log.Fatal(tx.Error)
	}
	for _, stat := range stats {
		db.Model(&stat).Update("weekday", int(stat.ScrapedAt.Weekday()))
	}
	r := setupRouter(db)
	r.Run()
}

func setupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.IndentedJSON(200, gin.H{
			"status": "ok",
		})
	})

	r.GET("/usage_stats", func(c *gin.Context) {
		var stats []models.UsageStat
		if tx := db.Find(&stats); tx.Error != nil {
			c.JSON(http.StatusInternalServerError, tx.Error)
			return
		}
		c.JSON(http.StatusOK, &stats)
	})
	return r
}
