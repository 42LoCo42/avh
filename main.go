package main

// #cgo LDFLAGS: -lpthread -lm

import (
	"bytes"
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path"
	"strings"

	_ "modernc.org/sqlite"
)

const (
	SECURE_PATH = "/secure"

	APP_DIR    = "app/"
	DB_PATH    = APP_DIR + "users.db"
	SRV_DIR    = APP_DIR + "srv/"
	SECURE_DIR = SRV_DIR + SECURE_PATH + "/"

	INIT_LENGTH      = 16
	ENV_JWT_KEY_NAME = "AVH_JWT_KEY"
)

//go:embed template.html
var videosTemplate []byte

var db *sql.DB
var jwtKey []byte

func main() {
	var err error

	jwtKeyString, ok := os.LookupEnv(ENV_JWT_KEY_NAME)
	if !ok {
		log.Fatalf("%s not set!", ENV_JWT_KEY_NAME)
	}
	jwtKey = []byte(jwtKeyString)

	init := false
	if _, err := os.Stat(DB_PATH); os.IsNotExist(err) {
		init = true
	}

	db, err = sql.Open("sqlite", DB_PATH)
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

	// admin stuff
	http.HandleFunc("/admin/listUsers", adminListUsers)
	http.HandleFunc("/admin/newUser", adminNewUser)
	http.HandleFunc("/admin/delUser", adminDelUser)
	http.HandleFunc("/admin/setUserPW", adminSetUserPW)
	http.HandleFunc("/admin/resetUserPW", adminResetUserPW)

	// user stuff
	http.HandleFunc("/changePW", userChangePW)
	http.HandleFunc("/login", userLogin)

	// specials
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		user, ok := userAuth(w, r)
		if !ok {
			return
		}

		if user != "admin" && !canUpload(user) {
			log.Printf("User %s may not upload!", user)
			noAuth(w, r, user)
			return
		}

		_, params, err := mime.ParseMediaType(r.Header.Get("content-type"))
		if err != nil {
			onErr(w, r, err)
			return
		}

		boundary, ok := params["boundary"]
		if !ok {
			badReq(w, r, "No boundary in multipart request!")
			return
		}

		reader := multipart.NewReader(r.Body, boundary)
		for {
			part, err := reader.NextPart()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					onErr(w, r, err)
					return
				}
			}

			if part.FormName() != "file" {
				continue
			}

			base := SECURE_DIR + path.Base(part.FileName())
			log.Printf("User %s uploads %s", user, base)

			file, err := os.Create(base)
			if err != nil {
				onErr(w, r, err)
				return
			}
			defer file.Close()

			if _, err := io.Copy(file, part); err != nil {
				onErr(w, r, err)
				return
			}

			http.ServeFile(w, r, SRV_DIR+"ok.html")
			return
		}

		badReq(w, r, "Could not process upload")

	})

	http.HandleFunc(SECURE_PATH, func(w http.ResponseWriter, r *http.Request) {
		user, ok := userAuth(w, r)
		if !ok {
			return
		}
		log.Printf("User %s got %s", user, r.URL.Path)

		var str strings.Builder

		videos, err := os.ReadDir(SECURE_DIR)
		if err != nil {
			onErr(w, r, err)
			return
		}

		for _, video := range videos {
			name := video.Name()
			log.Print(name)
			// if name == "thumbnails" || strings.HasSuffix(name, ".html") {
			// 	continue
			// }

			base := SECURE_PATH + "/" + name
			// poster := "/secure/thumbnails/" + strings.TrimSuffix(name, path.Ext(name)) + ".png"

			fmt.Fprintf(
				&str,
				`
<div>
	%s<br>
	<a href="%[2]s">Download</a><br>
	<video controls poster="%[3]s" preload="metadata">
		<source src="%[2]s">
	</video>
</div>
				`,
				name,
				base,
				"",
			)
		}

		final := []byte(str.String())
		w.Write(bytes.Replace(videosTemplate, []byte("PLACE_VIDEOS_HERE"), final, 1))
	})

	// static files
	fileServer := http.FileServer(http.Dir(SRV_DIR))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, SECURE_PATH) {
			user, ok := userAuth(w, r)
			if !ok {
				return
			}
			log.Printf("User %s got %s", user, r.URL.Path)

			if !strings.HasSuffix(r.URL.Path, ".html") {
				w.Header().Set("Content-Disposition", "attachment")
			}
		}

		if len(r.URL.Path) > 1 && strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, strings.TrimSuffix(r.URL.Path, "/"), http.StatusTemporaryRedirect)
		} else {
			fileServer.ServeHTTP(w, r)
		}
	})

	listener, err := net.Listen("tcp4", "0.0.0.0:8080")
	if err != nil {
		log.Fatal("could not create listener: ", err)
	}

	go func() {
		log.Print("Up and running!")
	}()

	log.Fatal(http.Serve(listener, nil))
}
