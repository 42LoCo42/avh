package main

import (
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"log"
	"math/big"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/pbkdf2"
)

const DB = "users.db"
const INIT_LENGTH = 16

var db *sql.DB

func main() {
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
	http.HandleFunc("/changePW", changePW)

	// static files
	http.Handle("/", http.FileServer(http.Dir("root")))

	log.Print("Up and running")
	http.ListenAndServe(":37812", nil)
}

func genHash(pass, salt string) string {
	return base64.StdEncoding.EncodeToString(
		pbkdf2.Key(
			[]byte(pass),
			[]byte(salt),
			10000,
			64,
			sha512.New,
		),
	)
}

// https://gist.github.com/denisbrodbeck/635a644089868a51eccd6ae22b2eb800
func GenerateRandomASCIIString(length int) (string, error) {
	result := ""
	for {
		if len(result) >= length {
			return result, nil
		}
		num, err := rand.Int(rand.Reader, big.NewInt(int64(127)))
		if err != nil {
			return "", err
		}
		n := num.Int64()
		// Make sure that the number/byte/letter is inside
		// the range of printable ASCII characters (excluding space and DEL)
		if n > 32 && n < 127 {
			result += string(rune(n))
		}
	}
}
