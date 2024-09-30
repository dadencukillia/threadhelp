package main

import (
	"crypto"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/registration"
)

var (
	email       = ""
	domain      = ""
	webrootPath = "/var/www/"
)

const (
	certDir            = "certs"
	certFilename       = "cert.crt"
	privateKeyFilename = "privkey.key"

	autoRenewInterval = time.Hour * 24 * 20 // Every 20 days
)

func loadFromEnv() error {
	email = strings.Trim(os.Getenv("EMAIL"), " ")
	if email == "" {
		return errors.New("failed to load env variable: EMAIL")
	}

	domain = strings.Trim(os.Getenv("DOMAIN"), " ")
	if domain == "" {
		return errors.New("failed to load env variable: DOMAIN")
	}
	if strings.Contains(domain, "http") || strings.Contains(domain, "/") || strings.Contains(domain, " ") {
		return errors.New("invalid env variable: DOMAIN")
	}

	webrootPath = strings.Trim(os.Getenv("WEBROOT"), " ")
	if webrootPath == "" {
		webrootPath = "/var/www/"
	}

	return nil
}

type User struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}
func (u User) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}
