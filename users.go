package main

import (
	"log"
	"net/http"
	"time"

	"github.com/aerogo/aero"
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

	_, err = db.Exec("insert into users values (?, ?, ?, false)", name, salt, hash)
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

func canUpload(name string) bool {
	query := db.QueryRow("select canUpload from users where name = ?", name)
	var canUpload bool
	if err := query.Scan(&canUpload); err != nil {
		log.Print(err)
		return false
	}

	return canUpload
}

func userChangePW(ctx aero.Context) error {
	form := ctx.Request().Internal().Form
	name := form.Get("name")
	oldPW := form.Get("oldPW")
	newPW := form.Get("newPW")
	repPW := form.Get("repPW")

	if newPW != repPW {
		return badReq(ctx, "Passwörter stimmen nicht überein!")
	}

	if !checkAuth(name, oldPW) {
		return noAuth(ctx, name)
	}

	if err := setPW(name, newPW); err != nil {
		return onErr(ctx, err)
	}

	log.Printf("User %s changed their password", name)
	return ctx.Redirect(http.StatusTemporaryRedirect, "ok.html?next=/")
}

func userAuth(ctx aero.Context) (string, error) {
	auth, err := ctx.Request().Internal().Cookie("auth")
	if err != nil {
		return "", noAuth(ctx, "")
	}

	user, ok := checkJWT(auth.Value)
	if !ok {
		return "", noAuth(ctx, user)
	}

	return user, nil
}

func userLogin(ctx aero.Context) error {
	form := ctx.Request().Internal().Form
	name := form.Get("name")
	pass := form.Get("pass")

	if !checkAuth(name, pass) {
		return noAuth(ctx, name)
	}

	token, err := issueJWT(name)
	if err != nil {
		return onErr(ctx, err)
	}

	log.Printf("User %s logged in", name)

	cookie := http.Cookie{
		Name:     "auth",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(ctx.Response().Internal(), &cookie)
	return ctx.Redirect(http.StatusTemporaryRedirect, "/secure")
}
