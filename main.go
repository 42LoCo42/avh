package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	"github.com/go-faster/errors"

	"github.com/glebarez/sqlite"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"

	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
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

		return Page(Index(videos)).Render(c.Response())
	})
	errors.Must(0, e.Start(":8080"))
}

type PageData struct {
	Title string
	Body  []g.Node
}

func Page(d PageData) g.Node {
	return c.HTML5(c.HTML5Props{
		Title:    d.Title,
		Language: "de",
		Body:     d.Body,
	})
}

func Login() PageData {
	title := "AvH-Videos - Login"
	return PageData{
		Title: title,
		Body: []g.Node{
			H1(g.Text(title)),
			Form(Action("/login"), Method("POST"),
				Input(Type("submit"), Style("display: none")),
				Table(
					Tr(
						Td(g.Text("Benutzername")),
						Td(Input(Type("text"), Name("username"))),
					),
					Tr(
						Td(g.Text("Passwort")),
						Td(Input(Type("password"), Name("password"))),
					),
				),
				Input(Type("submit"), Value("Login")),
			),
		},
	}
}

func Index(videos []string) PageData {
	style := strings.Join([]string{
		"display: flex",
		"flex-flow: column nowrap",
		"border: 1px solid black",
		"padding: 5px",
		"margin: 4px",
	}, "; ")

	title := "AvH-Videos"
	return PageData{
		Title: title,
		Body: []g.Node{
			H1(g.Text(title)),
			Div(Style("display: flex; flex-flow: row wrap"),
				g.Group(g.Map(videos, func(video string) g.Node {
					return Div(Style(style),
						Span(g.Text(video)),
						Video(
							Controls(),
							Preload("none"),
							Poster(fmt.Sprintf("%v.png", video)),
							Style("max-width: 100%; margin-top: 5px"),
							Source(Src(video)),
						),
					)
				})),
			),
		},
	}
}
