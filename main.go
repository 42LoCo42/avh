package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

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

	var err error
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

	// static files
	fileServer := http.FileServer(http.Dir("root"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		user, ok := userAuth(w, r)
		if !ok {
			return
		}

		log.Printf("User %s got %s", user, r.URL.Path)
		fileServer.ServeHTTP(w, r)
	})

	log.Print("Up and running")
	http.ListenAndServe(":37812", nil)
}
