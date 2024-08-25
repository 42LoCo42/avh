package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-faster/errors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
)

func main() {
	// init webserver
	e := echo.New()
	e.Use(
		middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format:           "${time_custom} ${remote_ip} - ${method} ${host}${uri} - ${status} ${error}\n",
			CustomTimeFormat: "2006/01/02 15:04:05",
		}),

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
