package main

import (
	"log"
	"os"
	"threadhelpServer/utils"
)

var logger = log.New(os.Stdout, "WEB | ", log.Ldate|log.Ltime|log.Lshortfile|log.Lmsgprefix)
var cacheStorage = utils.NewCacheStorage()

func main() {
	if err := utils.InitDB(&cacheStorage); err != nil {
		logger.Fatalln(err)
	}

	defer utils.CloseDB()

	logger.Fatalln(StartWebServer())
}
