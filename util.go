package main

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"log"
	"math/big"
	"net/http"

	"golang.org/x/crypto/pbkdf2"
)

func noAuth(w http.ResponseWriter, r *http.Request, name string) {
	if name == "" {
		log.Print("Login with no cookie!")
	} else {
		log.Printf("User %s failed login!", name)
	}

	http.Redirect(w, r, "/err.html?msg=Authentifizierung fehlgeschlagen", http.StatusTemporaryRedirect)
}

func badReq(w http.ResponseWriter, r *http.Request, msg string) {
	log.Print(msg)
	http.Redirect(w, r, "/err.html?msg="+msg, http.StatusTemporaryRedirect)
}

func onErr(w http.ResponseWriter, r *http.Request, err error) {
	log.Print(err)
	http.Redirect(w, r, "/err.html?msg"+err.Error(), http.StatusTemporaryRedirect)
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
