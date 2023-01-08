package main

import (
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/pbkdf2"
)

const DB = "users.db"
const INIT_LENGTH = 16

func main() {
	init := false
	if _, err := os.Stat(DB); os.IsNotExist(err) {
		init = true
	}

	db, err := sql.Open("sqlite3", "users.db")
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

		if err := newUser(db, "admin", pass); err != nil {
			log.Fatal(err)
		}

		log.Print("Created initial admin user with password ", pass)
	}

	http.HandleFunc("/admin/", func(w http.ResponseWriter, r *http.Request) {
		log.Print("admin login from ", r.RemoteAddr)
		name, pass, ok := r.BasicAuth()

		noAuth := func() {
			log.Printf("User %s failed admin login!", name)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		}

		badReq := func(msg string) {
			log.Print(msg)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(msg))
		}

		onErr := func(err error) {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		if !ok || name != "admin" || !checkAuth(db, name, pass) {
			noAuth()
			return
		}

		target := strings.TrimPrefix(r.URL.Path, "/admin/")

		switch target {
		case "listUsers":
			query, err := db.Query("select name from users")
			if err != nil {
				onErr(err)
				return
			}

			for query.Next() {
				var name string
				if err := query.Scan(&name); err != nil {
					onErr(err)
					return
				}

				fmt.Fprintln(w, name)
			}

		case "newUser":
			name := r.FormValue("name")
			if name == "" {
				badReq("name is blank")
				return
			}

			pass, err := GenerateRandomASCIIString(INIT_LENGTH)
			if err != nil {
				onErr(err)
				return
			}

			if err := newUser(db, name, pass); err != nil {
				onErr(err)
			}

			w.Write([]byte(pass))

		case "delUser":
			name := r.FormValue("name")
			if name == "" {
				badReq("name is blank")
				return
			}

			if _, err := db.Exec("delete from users where name = ?", name); err != nil {
				onErr(err)
				return
			}

		default:
			badReq("Unknown action")
		}
	})

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

func newUser(db *sql.DB, name, pass string) error {
	salt, err := GenerateRandomASCIIString(INIT_LENGTH)
	if err != nil {
		return nil
	}

	hash := genHash(pass, salt)
	_, err = db.Exec("insert into users values (?, ?, ?)", name, salt, hash)
	return err
}

func checkAuth(db *sql.DB, name, pass string) bool {
	query := db.QueryRow("select salt, hash from users where name = ?", name)
	var salt string
	var hash string
	if err := query.Scan(&salt, &hash); err != nil {
		log.Print(err)
		return false
	}

	return hash == genHash(pass, salt)
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
