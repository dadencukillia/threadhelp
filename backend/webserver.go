package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	midLogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var oauthAllowDomain = os.Getenv("OAUTH_ALLOW_DOMAIN")
var webpImageEncoding = os.Getenv("WEBP_IMAGE_ENCODING") == "true"
var sseClientsMutex = sync.Mutex{}
var sseClients = []*SSEClient{}

type SSEClient struct {
	PingingSkip uint32
	Mutex       sync.Mutex
	Queue       []string
}

func StartWebServer() error {
	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024,
	})

	app.Use(midLogger.New())
	app.Use(cors.New())

	apiGroup := app.Group("/api")
	apiGroup.Use(middlewareCheckGoogleAuth)

	apiGroup.Get("check", func(c fiber.Ctx) error {
		retInfo := map[string]any{}

		retInfo["admin"] = IsAdmin(c.Locals("email").(string))

		return c.Status(fiber.StatusOK).JSON(retInfo)
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
			if len(content) > 1024*1024*1024*20 {
				return c.SendStatus(fiber.StatusRequestEntityTooLarge)
			}

			attachedImages := []string{}
			contentImages := [][]byte{}

			ctx := &html.Node{
				Type:     html.ElementNode,
				DataAtom: atom.Div,
				Data:     "div",
			}
			nodes, err := html.ParseFragment(strings.NewReader(content), ctx)
			if err != nil {
				return c.SendStatus(fiber.StatusBadRequest)
			}

			for _, node := range nodes {
				ctx.AppendChild(node)
			}

			doc := goquery.NewDocumentFromNode(ctx)

			if !strings.Contains(content, "<img") && len(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(doc.Text(), " ", ""), "\n", ""), "\t", "")) < 5 {
				return c.SendStatus(fiber.StatusBadRequest)
			}

			for _, matched := range regexp.MustCompile("<\\s*([[:alpha:]]+)[ \\/>]").FindAllStringSubmatch(content, -1) {
				tagName := matched[1]
				if !slices.Contains(allowedTags, tagName) {
					return c.SendStatus(fiber.StatusBadRequest)
				}
			}

			doc.Find("img").Each(func(i int, s *goquery.Selection) {
				srcVal, ex := s.Attr("src")
				if !ex {
					s.Remove()
					return
				}

				s.SetAttr("alt", "Post image "+fmt.Sprint(i))
				ext := ""

				if strings.HasPrefix(srcVal, "data:image/jpeg;base64,") {
					ext = "jpg"
				} else if strings.HasPrefix(srcVal, "data:image/png;base64,") {
					ext = "png"
				} else if strings.HasPrefix(srcVal, "data:image/webp;base64,") {
					ext = "webp"
				}

				if ext != "" {
					uuidVal, err := uuid.NewRandom()
					for err != nil {
						uuidVal, err = uuid.NewRandom()
					}

					var name string
					if webpImageEncoding {
						name = uuidVal.String() + ".webp"
					} else {
						name = uuidVal.String() + "." + ext
					}

					base64Data := srcVal[strings.IndexRune(srcVal, ',')+1:]
					imgData, err := base64.StdEncoding.DecodeString(base64Data)
					if err != nil {
						s.Remove()
						return
					}

					// Converting png or jpeg file into webp, if WEBP_IMAGE_ENCODING is turned on
					if webpImageEncoding && ext != "webp" {
						options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 25)
						if err != nil {
							s.Remove()
							return
						}

						var img image.Image
						reader := bytes.NewReader(imgData)
						if ext == "jpg" {
							img, err = jpeg.Decode(reader)
						} else {
							img, err = png.Decode(reader)
						}

						if err != nil {
							s.Remove()
							return
						}

						buff := bytes.NewBuffer([]byte{})
						err = webp.Encode(buff, img, options)
						if err != nil {
							s.Remove()
							return
						}

						imgData = buff.Bytes()
					}

					attachedImages = append(attachedImages, name)
					contentImages = append(contentImages, imgData)
					s.SetAttr("src", "/images/"+name)
				}
			})

			content, err = doc.Html()
			if err != nil {
				logger.Println(err)
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

			userId := c.Locals("uid")
			email := c.Locals("email")
			displayName := c.Locals("displayName")
			var postId string
			var pubDate pgtype.Timestamp
			r := tx.QueryRow(
				DBCTX,
				"INSERT INTO posts(userId, userEmail, userDisplayName, content, attachedImages) VALUES($1, $2, $3, $4, $5) RETURNING id, pubDate",
				userId, email, displayName, content, strings.Join(attachedImages, ","),
			)
			if err := r.Scan(&postId, &pubDate); err != nil {
				logger.Println(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			if err := tx.Commit(DBCTX); err != nil {
				logger.Println(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			for i, imgPath := range attachedImages {
				os.WriteFile("images/"+imgPath, contentImages[i], 0666)
			}

			jsonData, err := json.Marshal(Post{
				ID:              postId,
				UserID:          userId.(string),
				UserDisplayName: displayName.(string),
				Content:         content,
				PubDate:         uint64(pubDate.Time.UnixMilli()),
			})
			if err != nil {
				return c.SendStatus(fiber.StatusCreated)
			}

			sseClientsMutex.Lock()
			defer sseClientsMutex.Unlock()
			for _, sc := range sseClients {
				sc.Mutex.Lock()
				sc.Queue = append(sc.Queue, "newPost;"+string(jsonData))
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

		var attachedImages string

		var row pgx.Row
		if IsAdmin(c.Locals("email").(string)) {
			row = tx.QueryRow(DBCTX, "DELETE FROM posts WHERE id=$1 RETURNING attachedImages", postId)
		} else {
			row = tx.QueryRow(DBCTX, "DELETE FROM posts WHERE userId=$1 AND id=$2 RETURNING attachedImages", userId, postId)
		}
		err = row.Scan(&attachedImages)
		if err != nil {
			logger.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		for _, img := range strings.Split(attachedImages, ",") {
			os.Remove("images/" + img)
		}

		if err := tx.Commit(DBCTX); err != nil {
			logger.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		sseClientsMutex.Lock()
		defer sseClientsMutex.Unlock()
		for _, sc := range sseClients {
			sc.Mutex.Lock()
			sc.Queue = append(sc.Queue, "delPost;"+postId)
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
				ID:              postId,
				UserID:          userId,
				UserDisplayName: userDisplayName,
				Content:         content,
				PubDate:         uint64(pubDate.Time.UnixMilli()),
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
				ID:              postId,
				UserID:          userId,
				UserDisplayName: userDisplayName,
				Content:         content,
				PubDate:         uint64(pubDate.Time.UnixMilli()),
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
				PingingSkip: 5,
				Mutex:       sync.Mutex{},
				Queue:       []string{},
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
					// Return true to disconnect

					msgs := []string{"PING"}
					clientQ.Mutex.Lock()
					defer clientQ.Mutex.Unlock()

					if len(clientQ.Queue) == 0 {
						if clientQ.PingingSkip == 0 {
							clientQ.PingingSkip = 5
						} else {
							clientQ.PingingSkip -= 1
							return false
						}
					} else {
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

	app.Get("/images/*", static.New(
		"./images",
		static.Config{
			Compress:      true,
			MaxAge:        int((5 * 24 * time.Hour).Seconds()),
			CacheDuration: 3 * time.Minute,
		},
	))
	app.Use(static.New(
		"./frontend",
		static.Config{
			MaxAge: int((2 * 24 * time.Hour).Seconds()),
		},
	))
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
		if err != nil {
			return false
		}

		if aEmail, ok := tokenInfo.Claims["email"]; ok {
			if strEmail, ok := aEmail.(string); ok {
				if strings.HasSuffix(strEmail, "@"+oauthAllowDomain) {
					if aDisplayName, ok := tokenInfo.Claims["name"]; ok {
						if strDisplayName, ok := aDisplayName.(string); ok {
							email = strEmail
							uid = tokenInfo.UID
							displayName = strDisplayName

							cacheStorage.SetCache("tokenInfo;"+authToken, map[string]string{
								"email":       email,
								"uid":         uid,
								"displayName": displayName,
							}, 3*time.Minute)

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
