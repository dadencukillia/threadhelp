package main

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DBADDRESS = os.Getenv("DBADDRESS")
var DBCTX = context.Background()
var db *pgxpool.Pool

type Post struct {
	ID string `json:"postId"`
	UserID string `json:"userId"`
	UserDisplayName string `json:"userDisplayName"`
	Content string `json:"post"`
	PubDate string `json:"pubTime"`
}

func InitDB() error {
	var err error
	db, err = pgxpool.New(DBCTX, DBADDRESS)

	if err == nil {
		logger.Println("DB connected")
	}

	return InitTables()
}

func InitTables() error {
	query := `
CREATE TABLE IF NOT EXISTS blacklist(
	gmail text PRIMARY KEY
);
CREATE TABLE IF NOT EXISTS posts(
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	userId text,
	userEmail text,
	userDisplayName text,
	content text,
	pubDate timestamp without time zone DEFAULT NOW(),
	attachedImages text
);
`
	con, err := db.Acquire(DBCTX)
	if err != nil {return err}
	defer con.Release()

	tx, err := con.Begin(DBCTX)
	if err != nil {return err}

	tx.Exec(DBCTX, query)
	return tx.Commit(DBCTX)
}

func CloseDB() {
	db.Close()
}

func IsInBlacklist(gmail string) bool {
	if cached, found := cacheStorage.GetCache("userBlacklist;" + gmail); found {
		if ret, ok := cached.(bool); ok {
			return ret
		}
	}

	var inBlacklist bool = false

	con, err := db.Acquire(DBCTX)
	if err != nil {return false}
	defer con.Release()

	if con.QueryRow(DBCTX, "SELECT EXISTS(SELECT 1 FROM blacklist WHERE gmail=$1 LIMIT 1)", gmail).Scan(&inBlacklist) != nil {
		return false
	}

	cacheStorage.SetCache("userBlacklist;" + gmail, inBlacklist, 10 * time.Minute)

	return inBlacklist
}
