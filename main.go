package main

// #cgo LDFLAGS: -lpthread -lm

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"os"
	"path"
	"strings"

	"github.com/aerogo/aero"
	_ "modernc.org/sqlite"
)

const DB = "users.db"
const INIT_LENGTH = 16
const ENV_JWT_KEY_NAME = "AVH_JWT_KEY"

var db *sql.DB
var jwtKey []byte

func main() {
	jwtKeyString, ok := os.LookupEnv(ENV_JWT_KEY_NAME)
	if !ok {
		log.Fatalf("%s not set!", ENV_JWT_KEY_NAME)
	}
	jwtKey = []byte(jwtKeyString)

	init := false
	if _, err := os.Stat(DB); os.IsNotExist(err) {
		init = true
	}

	videosTemplate, err := os.ReadFile("root/secure/template.html")
	if err != nil {
		log.Fatal(err)
	}

	db, err = sql.Open("sqlite", "users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if init {
		if _, err = db.Exec(`
			create table users (
				name text primary key,
				salt text not null,
				hash text not null,
				canUpload bool not null
			)`,
		); err != nil {
			log.Fatal(err)
		}

		pass, err := GenerateRandomASCIIString(INIT_LENGTH)
		if err != nil {
			log.Fatal(err)
		}

		if err := newUser("admin", pass); err != nil {
			log.Fatal(err)
		}

		log.Print("Created initial admin user with password ", pass)
	}

	if err := os.Chdir("root"); err != nil {
		log.Fatal(err)
	}

	app := aero.New()

	// user stuff
	app.Get("/changePW", userChangePW)
	app.Get("/login", userLogin)

	// specials
	app.Get("/upload", func(ctx aero.Context) error {
		user, err := userAuth(ctx)
		if err != nil {
			return err
		}

		if user != "admin" && !canUpload(user) {
			log.Printf("User %s may not upload!", user)
			return noAuth(ctx, user)
		}

		_, params, err := mime.ParseMediaType(ctx.Request().Header("content-type"))
		if err != nil {
			return onErr(ctx, err)
		}

		boundary, ok := params["boundary"]
		if !ok {
			return badReq(ctx, "No boundary in multipart request!")
		}

		reader := multipart.NewReader(ctx.Request().Body().Reader(), boundary)
		for {
			part, err := reader.NextPart()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return onErr(ctx, err)
				}
			}

			if part.FormName() != "file" {
				continue
			}

			base := "./root/secure/" + path.Base(part.FileName())
			log.Printf("User %s uploads %s", user, base)

			file, err := os.Create(base)
			if err != nil {
				return onErr(ctx, err)
			}
			defer file.Close()

			if _, err := io.Copy(file, part); err != nil {
				return onErr(ctx, err)
			}

			return ctx.File("root/ok.html")
		}

		return badReq(ctx, "Could not process upload")
	})

	app.Get("/secure", func(ctx aero.Context) error {
		user, err := userAuth(ctx)
		if err != nil {
			return err
		}
		log.Printf("User %s got %s", user, ctx.Request().Internal().URL.Path)

		var str strings.Builder

		videos, err := os.ReadDir("root/secure")
		if err != nil {
			return onErr(ctx, err)
		}

		for _, video := range videos {
			name := video.Name()
			if name == "thumbnails" || strings.HasSuffix(name, ".html") {
				continue
			}

			base := "/secure/" + name

			fmt.Fprintf(
				&str,
				`
<div>
	%s<br>
	<a href="%[2]s">Download</a><br>
	<video controls preload="metadata">
		<source src="%[2]s">
	</video>
</div>
				`,
				name,
				base,
			)
		}

		final := str.String()
		return ctx.HTML(strings.Replace(string(videosTemplate), "PLACE_VIDEOS_HERE", final, 1))
	})

	// static files
	app.Get("/*file", func(ctx aero.Context) error {
		path := ctx.Request().Internal().URL.Path

		if strings.HasPrefix(path, "/secure/") {
			user, err := userAuth(ctx)
			if err != nil {
				return err
			}
			log.Printf("User %s got %s", user, path)

			if !strings.HasSuffix(path, ".html") {
				ctx.Response().SetHeader("content-disposition", "attachment")
			}
		}

		return ctx.File(ctx.Get("file"))
	})

	log.Print("Up and running")

	// go func() {
	// 	var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	// 		http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusTemporaryRedirect)
	// 	}

	// 	upgrader := &http.Server{
	// 		Addr:    ":80",
	// 		Handler: handler,
	// 	}
	// 	log.Fatal(upgrader.ListenAndServe())
	// }()

	app.Config.Ports.HTTP = 37812
	app.Run()
	// log.Fatal(http.ListenAndServeTLS(":443", "cert", "key", nil))
}
