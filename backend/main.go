package main

import (
	"context"
	"log"
	"os"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var logger = log.New(os.Stdout, "WEB | ", log.Ldate|log.Ltime|log.Lshortfile|log.Lmsgprefix)
var cacheStorage = NewCacheStorage()
var firebaseAuth *auth.Client

func main() {
	if err := InitFirebase(); err != nil {
		logger.Fatalln(err)
	}

	if err := InitDB(); err != nil {
		logger.Fatalln(err)
	}

	defer CloseDB()

	logger.Fatalln(StartWebServer())
}

func InitFirebase() error {
	opt := option.WithCredentialsFile("./firebaseSecretKey.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}

	firebaseAuth, err = app.Auth(context.Background())
	if err != nil {
		return err
	}

	return nil
}
