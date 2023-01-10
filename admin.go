package main

import (
	"fmt"
	"log"
	"net/http"
)

func adminLoginCheck(w http.ResponseWriter, r *http.Request) bool {
	log.Print("admin login from ", r.RemoteAddr)
	name, pass, ok := r.BasicAuth()

	if !ok || name != "admin" || !checkAuth(name, pass) {
		noAuth(w, r, name)
		return false
	}

	return true
}

func adminListUsers(w http.ResponseWriter, r *http.Request) {
	if !adminLoginCheck(w, r) {
		return
	}

	query, err := db.Query("select name from users")
	if err != nil {
		onErr(w, r, err)
		return
	}

	for query.Next() {
		var name string
		if err := query.Scan(&name); err != nil {
			onErr(w, r, err)
			return
		}

		fmt.Fprintln(w, name)
	}
}

func adminNewUser(w http.ResponseWriter, r *http.Request) {
	if !adminLoginCheck(w, r) {
		return
	}

	name := r.FormValue("name")
	if name == "" {
		badReq(w, r, "name is blank")
		return
	}

	pass, err := GenerateRandomASCIIString(INIT_LENGTH)
	if err != nil {
		onErr(w, r, err)
		return
	}

	if err := newUser(name, pass); err != nil {
		onErr(w, r, err)
	}

	w.Write([]byte(pass))
}

func adminDelUser(w http.ResponseWriter, r *http.Request) {
	if !adminLoginCheck(w, r) {
		return
	}

	name := r.FormValue("name")
	if name == "" {
		badReq(w, r, "name is blank")
		return
	}

	if _, err := db.Exec("delete from users where name = ?", name); err != nil {
		onErr(w, r, err)
		return
	}
}

func adminSetUserPW(w http.ResponseWriter, r *http.Request) {
	if !adminLoginCheck(w, r) {
		return
	}

	name := r.FormValue("name")
	pass := r.FormValue("pass")

	if err := setPW(name, pass); err != nil {
		onErr(w, r, err)
		return
	}
}

func adminResetUserPW(w http.ResponseWriter, r *http.Request) {
	if !adminLoginCheck(w, r) {
		return
	}

	name := r.FormValue("name")
	if name == "" {
		badReq(w, r, "name is blank")
		return
	}

	pass, err := resetPW(name)
	if err != nil {
		onErr(w, r, err)
		return
	}

	w.Write([]byte(pass))
}
