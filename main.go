package main

// #cgo LDFLAGS: -lpthread -lm

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
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

	db, err = sql.Open("sqlite3", "users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if init {
		if _, err = db.Exec(
			"create table users (name text primary key, salt text not null, hash text not null)",
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

	// video list
	http.HandleFunc("/secure", func(w http.ResponseWriter, r *http.Request) {
		user, ok := userAuth(w, r)
		if !ok {
			return
		}
		log.Printf("User %s got %s", user, r.URL.Path)

		var str strings.Builder

		videos, err := os.ReadDir("root/secure")
		if err != nil {
			onErr(w, r, err)
			return
		}

		for _, video := range videos {
			name := video.Name()
			if name == "template.html" {
				continue
			}

			fmt.Fprintf(
				&str,
				`
<div>
	%s&emsp;<a href="%[2]s">Download</a><br>
	<video controls>
		<source src="%[2]s">
	</video>
</div>
				`,
				name,
				"/secure/" + name,
			)
		}

		final := []byte(str.String())
		w.Write(bytes.Replace(videosTemplate, []byte("PLACE_VIDEOS_HERE"), final, 1))
	})

	// static files
	fileServer := http.FileServer(http.Dir("root"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/secure/") {
			user, ok := userAuth(w, r)
			if !ok {
				return
			}
			log.Printf("User %s got %s", user, r.URL.Path)
			w.Header().Set("Content-Disposition", "attachment")
		}


		fileServer.ServeHTTP(w, r)
	})

	log.Print("Up and running")
	http.ListenAndServe(":37812", nil)
}
