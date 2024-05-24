package main

import (
	"crypto/rand"
	"os"
	"strings"

	"github.com/42LoCo42/avh/jade"
	"github.com/go-faster/errors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	Name      string `gorm:"primaryKey"`
	Salt      string
	Hash      string
	CanUpload bool `gorm:"column:canUpload"`
}

func main() {
	// init db
	db := errors.Must(gorm.Open(sqlite.Open("users.db")))
	errors.Must(0, db.AutoMigrate(&User{}))

	// init JWT key
	jwtKey := make([]byte, 64)
	errors.Must(rand.Read(jwtKey))

	// init webserver
	e := echo.New()
	e.Use(
		middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format:           "${time_custom} ${remote_ip} - ${method} ${host}${uri} - ${status} ${error}\n",
			CustomTimeFormat: "2006/01/02 15:04:05",
		}),

		Auth(e, db, jwtKey),

		middleware.Static("videos"),
	)
	e.GET("/", func(c echo.Context) error {
		entries, err := os.ReadDir("videos")
		if err != nil {
			return errors.Wrap(err, "could not list videos")
		}

		videos := []string{}
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".png") {
				videos = append(videos, name)
			}
		}

		jade.Jade_index(videos, c.Response())
		return nil
	})
	errors.Must(0, e.Start(":8080"))
}
