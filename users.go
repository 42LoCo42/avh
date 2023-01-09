package main

import (
	"log"
	"net/http"
)

func newHash(pass string) (string, string, error) {
	salt, err := GenerateRandomASCIIString(INIT_LENGTH)
	if err != nil {
		return "", "", err
	}

	return salt, genHash(pass, salt), nil
}

func newUser(name, pass string) error {
	salt, hash, err := newHash(pass)
	if err != nil {
		return err
	}

	_, err = db.Exec("insert into users values (?, ?, ?)", name, salt, hash)
	return err
}

func setPW(name, pass string) error {
	salt, hash, err := newHash(pass)
	if err != nil {
		return err
	}

	if _, err := db.Exec(
		"update users set salt = ?, hash = ? where name = ?",
		salt, hash, name,
	); err != nil {
		return err
	}

	return nil
}

func resetPW(name string) (string, error) {
	pass, err := GenerateRandomASCIIString(INIT_LENGTH)
	if err != nil {
		return "", err
	}

	if err := setPW(name, pass); err != nil {
		return "", err
	}

	return pass, nil
}

func checkAuth(name, pass string) bool {
	query := db.QueryRow("select salt, hash from users where name = ?", name)
	var salt string
	var hash string
	if err := query.Scan(&salt, &hash); err != nil {
		log.Print(err)
		return false
	}

	return hash == genHash(pass, salt)
}

func changePW(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	oldPW := r.FormValue("oldPW")
	newPW := r.FormValue("newPW")
	repPW := r.FormValue("repPW")

	if newPW != repPW {
		badReq(w, "Passwords don't match!")
		return
	}

	if !checkAuth(name, oldPW) {
		noAuth(w, name)
		return
	}

	if err := setPW(name, newPW); err != nil {
		onErr(w, err)
		return
	}

	log.Printf("User %s changed their password", name)
}
