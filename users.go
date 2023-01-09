package main

import (
	"log"
	"net/http"
	"time"
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

func userChangePW(w http.ResponseWriter, r *http.Request) {
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

func userAuth(w http.ResponseWriter, r *http.Request) (string, bool) {
	auth, err := r.Cookie("auth")
	if err != nil {
		noAuth(w, "")
		return "", false
	}

	user, ok := checkJWT(auth.Value)
	if !ok {
		noAuth(w, user)
		return "", false
	}

	return user, true
}

func userLogin(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	pass := r.FormValue("pass")

	if !checkAuth(name, pass) {
		noAuth(w, name)
		return
	}

	token, err := issueJWT(name)
	if err != nil {
		onErr(w, err)
		return
	}

	log.Printf("User %s logged in", name)

	cookie := http.Cookie{
		Name:     "auth",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
}
