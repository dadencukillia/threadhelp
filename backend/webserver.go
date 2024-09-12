package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	midLogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/valyala/fasthttp"
)

var oauthAllowDomain = os.Getenv("OAUTHALLOWDOMAIN")
var sseClientsMutex = sync.Mutex{}
var sseClients = []*SSEClient{}

type SSEClient struct {
	Mutex sync.Mutex
	Queue []string
}

func pgtimeToString(pgTime pgtype.Timestamp) string {
	zone, _ := time.LoadLocation("Europe/Kiev")
	return pgTime.Time.In(zone).Format("2006.01.02 15:04")
}

func StartWebServer() error {
	app := fiber.New()

	app.Use(midLogger.New())
	app.Use(cors.New())

	apiGroup := app.Group("/api")
	apiGroup.Use(middlewareCheckGoogleAuth)

	apiGroup.Get("check", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	apiGroup.Post("sendPost", func(c fiber.Ctx) error {
		allowedTags := []string{
			"p", "img", "strong", "a", "em", "u", "pre", "span",
		}

		var body map[string]string
		if json.Unmarshal(c.Body(), &body) != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		if content, ok := body["content"]; ok {
			for _, matched := range regexp.MustCompile("<([[:alpha:]]+)[ \\/>]").FindAllStringSubmatch(content, -1) {
				tagName := matched[1]
				if !slices.Contains(allowedTags, tagName) {
					return c.SendStatus(fiber.StatusBadRequest)
				}
			}

			if strings.Count(content, "<") != strings.Count(content, ">") {
				return c.SendStatus(fiber.StatusBadRequest)
			}

			con, err := db.Acquire(DBCTX)
			if err != nil {
				logger.Println(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			defer con.Release()

			tx, err := con.Begin(DBCTX)
			if err != nil {
				logger.Println(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			userId := c.Locals("uid")
			email := c.Locals("email")
			displayName := c.Locals("displayName")
			var postId string
			var pubDate pgtype.Timestamp
			r := tx.QueryRow(DBCTX, "INSERT INTO posts(userId, userEmail, userDisplayName, content) VALUES($1, $2, $3, $4) RETURNING id, pubDate", userId, email, displayName, content)
			if err := r.Scan(&postId, &pubDate); err != nil {
				logger.Println(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			if err := tx.Commit(DBCTX); err != nil {
				logger.Println(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			jsonData, err := json.Marshal(Post {
				ID: postId,
				UserID: userId.(string),
				UserDisplayName: displayName.(string),
				Content: content,
				PubDate: pgtimeToString(pubDate),
			})
			if err != nil {
				return c.SendStatus(fiber.StatusCreated)
			}

			sseClientsMutex.Lock()
			defer sseClientsMutex.Unlock()
			for _, sc := range sseClients {
				sc.Mutex.Lock()
				sc.Queue = append(sc.Queue, "newPost;" + string(jsonData))
				sc.Mutex.Unlock()
			}

			return c.Status(fiber.StatusOK).SendString(postId)
		}

		return c.SendStatus(fiber.StatusBadRequest)
	})

	apiGroup.Post("deletePost", func(c fiber.Ctx) error {
		var body map[string]string
		if json.Unmarshal(c.Body(), &body) != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		postId, ok := body["id"]
		if !ok {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		userId, ok := c.Locals("uid").(string)
		if !ok {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		con, err := db.Acquire(DBCTX)
		if err != nil {
			logger.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		defer con.Release()

		tx, err := con.Begin(DBCTX)
		if err != nil {
			logger.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		_, err = tx.Exec(DBCTX, "DELETE FROM posts WHERE userId=$1 AND id=$2;", userId, postId)
		if err != nil {
			logger.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if err := tx.Commit(DBCTX); err != nil {
			logger.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		sseClientsMutex.Lock()
		defer sseClientsMutex.Unlock()
		for _, sc := range sseClients {
			sc.Mutex.Lock()
			sc.Queue = append(sc.Queue, "delPost;" + postId)
			sc.Mutex.Unlock()
		}

		return c.SendStatus(fiber.StatusOK)
	})

	apiGroup.Get("tenNewestPosts", func(c fiber.Ctx) error {
		con, err := db.Acquire(DBCTX)
		if err != nil {
			logger.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		defer con.Release()

		rows, err := con.Query(DBCTX, "SELECT id, pubDate, userId, userDisplayName, content FROM posts ORDER BY pubDate DESC LIMIT 10")
		defer rows.Close()
		if err != nil {
			logger.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		ret := []Post{}

		for rows.Next() {
			var postId, userId, userDisplayName, content string
			var pubDate pgtype.Timestamp
			err := rows.Scan(&postId, &pubDate, &userId, &userDisplayName, &content)
			if err != nil {
				logger.Println(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			ret = append(ret, Post{
				ID: postId,
				UserID: userId,
				UserDisplayName: userDisplayName,
				Content: content,
				PubDate: pgtimeToString(pubDate),
			})
		}


		return c.Status(fiber.StatusOK).JSON(ret)
	})

	apiGroup.Get("getNextTenPosts/:postId", func(c fiber.Ctx) error {
		postId := c.Params("postId", "")
		if postId == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		con, err := db.Acquire(DBCTX)
		if err != nil {
			logger.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		defer con.Release()

		rows, err := con.Query(DBCTX, "SELECT id, pubDate, userId, userDisplayName, content FROM posts WHERE pubDate < (SELECT pubDate FROM posts WHERE id=$1 LIMIT 1) ORDER BY pubDate DESC LIMIT 10", postId)
		defer rows.Close()
		if err != nil {
			logger.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		ret := []Post{}

		for rows.Next() {
			var postId, userId, userDisplayName, content string
			var pubDate pgtype.Timestamp
			err := rows.Scan(&postId, &pubDate, &userId, &userDisplayName, &content)
			if err != nil {
				logger.Println(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			ret = append(ret, Post{
				ID: postId,
				UserID: userId,
				UserDisplayName: userDisplayName,
				Content: content,
				PubDate: pgtimeToString(pubDate),
			})
		}

		return c.Status(fiber.StatusOK).JSON(ret)
	})

	apiGroup.Get("/events", func(c fiber.Ctx) error {
		ctx := c.Context()

		ctx.SetContentType("text/event-stream")
		ctx.Response.Header.Set("Cache-Control", "no-cache")
		ctx.Response.Header.Set("Connection", "keep-alive")
		ctx.Response.Header.Set("Transfer-Encoding", "chunked")
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Cache-Control")
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
		ctx.SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			sseClientsMutex.Lock()
			clientQ := SSEClient{
				Mutex: sync.Mutex{},
				Queue: []string{},
			}
			sseClients = append(sseClients, &clientQ)
			sseClientsMutex.Unlock()

			defer func() {
				sseClientsMutex.Lock()
				defer sseClientsMutex.Unlock()
				for i, c := range sseClients {
					if c == &clientQ {
						sseClients = append(sseClients[:i], sseClients[i+1:]...)
						return
					}
				}
				ctx.Done()
			}()

			for {
				if func() bool {
					msgs := []string{"PING"}
					clientQ.Mutex.Lock()
					defer clientQ.Mutex.Unlock()

					if len(clientQ.Queue) != 0 {
						msgs = clientQ.Queue
					}

					for _, msg := range msgs {
						n, err := fmt.Fprintf(w, "data: %s\n\n", msg)
						if err != nil {
							logger.Println(err)
							return true
						}
						if n == 0 {
							logger.Println("wrong n=0")
							return true
						}

						if err := w.Flush(); err != nil {
							logger.Println(err)
							return true
						}
					}
					
					clientQ.Queue = []string{}

					return false
				}() {
					return
				}

				time.Sleep(1 * time.Second)
			}
		}))

		return nil
	})

	apiGroup.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotFound)
	})

	app.Use(static.New("./frontend"))
	app.Use(func(c fiber.Ctx) error {
		c.Context().SetContentType("text/html")
		return c.Status(fiber.StatusOK).SendFile("./frontend/index.html")
	})

	return app.Listen(":80")
}

func middlewareCheckGoogleAuth(c fiber.Ctx) error {
	var email, uid, displayName string

	if !func() bool {
		authToken := c.Get("Auth-Token", "")
		if len(authToken) == 0 {
			return false
		}

		if tokenInfo, ok := cacheStorage.GetCache("tokenInfo;" + authToken); ok {
			if mapTokenInfo, ok := tokenInfo.(map[string]string); ok {
				email = mapTokenInfo["email"]
				uid = mapTokenInfo["uid"]
				displayName = mapTokenInfo["displayName"]

				return true
			}
		}

		tokenInfo, err := firebaseAuth.VerifyIDTokenAndCheckRevoked(context.Background(), authToken)
		if err != nil {return false}

		if aEmail, ok := tokenInfo.Claims["email"]; ok {
			if strEmail, ok := aEmail.(string); ok {
				if strings.HasSuffix(strEmail, "@" + oauthAllowDomain) {
					if aDisplayName, ok := tokenInfo.Claims["name"]; ok {
						if strDisplayName, ok := aDisplayName.(string); ok {
							email = strEmail
							uid = tokenInfo.UID
							displayName = strDisplayName

							cacheStorage.SetCache("tokenInfo;" + authToken, map[string]string{
								"email": email,
								"uid": uid,
								"displayName": displayName,
							}, 3 * time.Minute)

							return true
						}
					}
				}
			}
		}

		return false
	}() {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	c.Locals("email", email)
	c.Locals("uid", uid)
	c.Locals("displayName", displayName)
	
	if IsInBlacklist(email) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	return c.Next()
}