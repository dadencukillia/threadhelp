package utils

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DBADDRESS = os.Getenv("DB_ADDRESS")
var DBCTX = context.Background()
var cacheStorage *CacheStorage
var db *pgxpool.Pool

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprint(time.Time(t).UnixMilli())), nil
}

type Post struct {
	ID              string   `json:"postId"`
	UserID          string   `json:"userId,omitempty"`
	UserDisplayName string   `json:"userDisplayName,omitempty"`
	PubDate         JSONTime `json:"pubTime,omitempty"`
	Content         string	`json:"-"`
	UserEmail       string `json:"-"`
	AttachedImages  []string `json:"-"`
}

func InitDB(cs *CacheStorage) error {
	cacheStorage = cs

	var err error
	db, err = pgxpool.New(DBCTX, DBADDRESS)

	if err != nil {
		return err
	}

	return InitTables()
}

func InitTables() error {
	query := `
CREATE TABLE IF NOT EXISTS blacklist(
	gmail text PRIMARY KEY
);
CREATE TABLE IF NOT EXISTS admins(
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
	if err != nil {
		return err
	}
	defer con.Release()

	tx, err := con.Begin(DBCTX)
	if err != nil {
		return err
	}

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
	if err != nil {
		return false
	}
	defer con.Release()

	if con.QueryRow(DBCTX, "SELECT EXISTS(SELECT 1 FROM blacklist WHERE gmail=$1 LIMIT 1)", gmail).Scan(&inBlacklist) != nil {
		return false
	}

	cacheStorage.SetCache("userBlacklist;"+gmail, inBlacklist, 10*time.Minute)

	return inBlacklist
}

func IsAdmin(gmail string) bool {
	if cached, found := cacheStorage.GetCache("userAdmin;" + gmail); found {
		if ret, ok := cached.(bool); ok {
			return ret
		}
	}

	var isAdmin bool = false

	con, err := db.Acquire(DBCTX)
	if err != nil {
		return false
	}
	defer con.Release()

	if con.QueryRow(DBCTX, "SELECT EXISTS(SELECT 1 FROM admins WHERE gmail=$1 LIMIT 1)", gmail).Scan(&isAdmin) != nil {
		return false
	}

	cacheStorage.SetCache("userAdmin;"+gmail, isAdmin, 10*time.Minute)

	return isAdmin
}

func AddPost(post Post) (Post, error) {
	con, err := db.Acquire(DBCTX)
	if err != nil {
		return Post{}, err
	}
	defer con.Release()

	tx, err := con.Begin(DBCTX)
	if err != nil {
		return Post{}, err
	}

	var pubDate pgtype.Timestamp

	r := tx.QueryRow(
		DBCTX,
		"INSERT INTO posts(userId, userEmail, userDisplayName, content, attachedImages) VALUES($1, $2, $3, $4, $5) RETURNING id, pubDate",
		post.UserID, post.UserEmail, post.UserDisplayName, post.Content, strings.Join(post.AttachedImages, ","),
	)
	if err := r.Scan(&post.ID, &pubDate); err != nil {
		return Post{}, err
	}

	post.PubDate = JSONTime(pubDate.Time)

	if err := tx.Commit(DBCTX); err != nil {
		return Post{}, err
	}

	return post, nil
}

func DeletePost(postId string, userId string) (attachedImages []string, error error) {
	con, err := db.Acquire(DBCTX)
	if err != nil {
		return []string{}, err
	}
	defer con.Release()

	tx, err := con.Begin(DBCTX)
	if err != nil {
		return []string{}, err
	}

	var attachedImgs string
	row := tx.QueryRow(DBCTX, "DELETE FROM posts WHERE userId=$1 AND id=$2 RETURNING attachedImages", userId, postId)

	err = row.Scan(&attachedImgs)
	if err != nil {
		return []string{}, err
	}

	if err := tx.Commit(DBCTX); err != nil {
		return []string{}, err
	}

	return strings.Split(attachedImgs, ","), nil
}

func DeletePostAdmin(postId string) (attachedImages []string, error error) {
	con, err := db.Acquire(DBCTX)
	if err != nil {
		return []string{}, err
	}
	defer con.Release()

	tx, err := con.Begin(DBCTX)
	if err != nil {
		return []string{}, err
	}

	var attachedImgs string
	row := tx.QueryRow(DBCTX, "DELETE FROM posts WHERE id=$1 RETURNING attachedImages", postId)

	err = row.Scan(&attachedImgs)
	if err != nil {
		return []string{}, err
	}

	if err := tx.Commit(DBCTX); err != nil {
		return []string{}, err
	}

	return strings.Split(attachedImgs, ","), nil
}

func GetNewestPosts(count uint32) ([]Post, error) {
	con, err := db.Acquire(DBCTX)
	if err != nil {
		return []Post{}, err
	}
	defer con.Release()

	rows, err := con.Query(DBCTX, "SELECT id, pubDate, userId, userDisplayName FROM posts ORDER BY pubDate DESC LIMIT $1", count)
	if err != nil {
		return []Post{}, err
	}

	defer rows.Close()

	posts := []Post{}
	for rows.Next() {
		var postId, userId, userDisplayName string
		var pubDate pgtype.Timestamp
		err := rows.Scan(&postId, &pubDate, &userId, &userDisplayName)
		if err != nil {
			return []Post{}, err
		}

		posts = append(posts, Post{
			ID:              postId,
			UserID:          userId,
			UserDisplayName: userDisplayName,
			PubDate:         JSONTime(pubDate.Time),
		})
	}

	return posts, nil
}

func GetNewestPostsFrom(postId string, count uint32) ([]Post, error) {
	con, err := db.Acquire(DBCTX)
	if err != nil {
		return []Post{}, err
	}
	defer con.Release()

	rows, err := con.Query(DBCTX, "SELECT id, pubDate, userId, userDisplayName FROM posts WHERE pubDate < (SELECT pubDate FROM posts WHERE id=$1 LIMIT 1) ORDER BY pubDate DESC LIMIT $2", postId, count)
	if err != nil {
		return []Post{}, err
	}

	defer rows.Close()

	posts := []Post{}
	for rows.Next() {
		var postId, userId, userDisplayName string
		var pubDate pgtype.Timestamp
		err := rows.Scan(&postId, &pubDate, &userId, &userDisplayName)
		if err != nil {
			return []Post{}, err
		}

		posts = append(posts, Post{
			ID:              postId,
			UserID:          userId,
			UserDisplayName: userDisplayName,
			PubDate:         JSONTime(pubDate.Time),
		})
	}

	return posts, nil
}

func GetPostContent(postId string) (string, error) {
	con, err := db.Acquire(DBCTX)
	if err != nil {
		return "", err
	}
	defer con.Release()

	var content string

	row := con.QueryRow(DBCTX, "SELECT content FROM posts WHERE id=$1", postId)
	if err := row.Scan(&content); err != nil {
		return "", err
	}

	return content, nil
}
