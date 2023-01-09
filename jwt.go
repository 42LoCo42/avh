package main

import (
	"log"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

func issueJWT(name string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   name,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	ss, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return ss, nil
}

func checkJWT(token string) (string, bool) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		log.Print("checkJWT error: ", err)
		return "", false
	}

	return claims.Subject, claims.ExpiresAt.After(time.Now())
}
