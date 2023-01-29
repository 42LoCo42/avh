package main

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"log"
	"math/big"
	"net/http"

	"github.com/aerogo/aero"
	"golang.org/x/crypto/pbkdf2"
)

func noAuth(ctx aero.Context, name string) error {
	if name == "" {
		log.Print("Login with no cookie!")
	} else {
		log.Printf("User %s failed login!", name)
	}

	return ctx.Redirect(http.StatusTemporaryRedirect, "/err.html?msg=Authentifizierung fehlgeschlagen")
}

func badReq(ctx aero.Context, msg string) error {
	log.Print(msg)
	return ctx.Redirect(http.StatusTemporaryRedirect, "/err.html?msg="+msg)
}

func onErr(ctx aero.Context, err error) error {
	log.Print(err)
	return ctx.Redirect(http.StatusTemporaryRedirect, "/err.html?msg"+err.Error())
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
