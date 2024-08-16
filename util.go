package main

import (
	"crypto/sha512"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/go-faster/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"
)

func FindUser(db *gorm.DB, username string) (*User, error) {
	var user User
	if err := db.Where("name = ?", username).First(&user).Error; err != nil {
		return nil, errors.Wrapf(err, "user %v not found", username)
	}
	return &user, nil
}

func MkHash(pass, salt string) string {
	return base64.StdEncoding.EncodeToString(
		pbkdf2.Key([]byte(pass), []byte(salt), 10000, 64, sha512.New))
}

func MkCookie(user string, jwtKey []byte) (*http.Cookie, error) {
	now := time.Now()
	exp := now.Add(time.Hour * 24)
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(exp),
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    "avh",
		Subject:   user,
	})

	signed, err := token.SignedString(jwtKey)
	if err != nil {
		return nil, errors.Wrap(err, "could not sign auth token")
	}

	return &http.Cookie{
		Name:     "auth",
		Value:    signed,
		Path:     "/",
		Expires:  exp,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}, nil
}

func Auth(e *echo.Echo, db *gorm.DB, jwtKey []byte) echo.MiddlewareFunc {
	fail := func(c echo.Context, err error) error {
		resp := c.Response()
		resp.Status = http.StatusUnauthorized
		return Page(Login()).Render(c.Response())
	}

	login := e.POST("/login", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		user, err := FindUser(db, username)
		if err != nil {
			return fail(c, err)
		}

		hash := MkHash(password, user.Salt)
		if hash != user.Hash {
			return fail(c, errors.New("hash mismatch"))
		}

		cookie, err := MkCookie(username, jwtKey)
		if err != nil {
			return fail(c, errors.Wrap(err, "could not create session cookie"))
		}

		c.SetCookie(cookie)
		log.Printf("\x1B[1;32mUser %v logged in!\x1B[m", username)
		return c.Redirect(http.StatusSeeOther, "/")
	})

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().URL.Path == login.Path {
				return next(c)
			}

			auth, err := c.Cookie("auth")
			if err != nil {
				return fail(c, errors.Wrap(err, "could not get auth cookie"))
			}

			token, err := jwt.ParseWithClaims(
				auth.Value,
				&jwt.RegisteredClaims{},
				func(t *jwt.Token) (interface{}, error) {
					return jwtKey, nil
				},
				jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Name}),
				jwt.WithIssuer("avh"),
			)
			if err != nil {
				return fail(c, errors.Wrap(err, "invalid auth token"))
			}

			username := token.Claims.(*jwt.RegisteredClaims).Subject
			user, err := FindUser(db, username)
			if err != nil {
				return fail(c, errors.Wrapf(err, "auth token has invalid user %v", username))
			}

			c.Set("user", user)
			return next(c)
		}
	}
}
