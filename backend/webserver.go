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
	"log"
	"net/http"
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
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var oauthAllowDomain = os.Getenv("OAUTH_ALLOW_DOMAIN")
var webpImageEncoding = os.Getenv("WEBP_IMAGE_ENCODING") == "true"
var httpsDomain = os.Getenv("HTTPS_DOMAIN")
var useHttps = os.Getenv("USE_HTTPS") == "true"
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

			post, err := AddPost(Post{
				UserID:          c.Locals("uid").(string),
				userEmail:       c.Locals("email").(string),
				UserDisplayName: c.Locals("displayName").(string),
				Content:         content,
				attachedImages:  attachedImages,
			})
			if err != nil {
				logger.Println(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			for i, imgPath := range attachedImages {
				os.WriteFile("images/"+imgPath, contentImages[i], 0666)
			}

			jsonData, err := json.Marshal(post)
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

			return c.Status(fiber.StatusOK).SendString(post.ID)
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

		isAdmin := IsAdmin(c.Locals("email").(string))

		var attachedImages []string
		var err error

		if isAdmin {
			attachedImages, err = DeletePostAdmin(postId)
		} else {
			attachedImages, err = DeletePost(postId, userId)
		}

		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		for _, img := range attachedImages {
			os.Remove("images/" + img)
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
		posts, err := GetNewestPosts(10)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.Status(fiber.StatusOK).JSON(posts)
	})

	apiGroup.Get("getNextTenPosts/:postId", func(c fiber.Ctx) error {
		postId := c.Params("postId", "")
		if postId == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		posts, err := GetNewestPostsFrom(postId, 10)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.Status(fiber.StatusOK).JSON(posts)
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

	if useHttps {
		go runHttpRedirector()

		for {
			time.Sleep(10 * time.Second)
			logger.Println(app.Listen(":443", fiber.ListenConfig{
				CertFile:    "/etc/letsencrypt/cert.crt",
				CertKeyFile: "/etc/letsencrypt/privkey.key",
			}))
		}
	} else {
		return app.Listen(":80")
	}
}

func runHttpRedirector() {
	http.Handle("GET /.well-known/acme-challenge/", http.StripPrefix("/.well-known/acme-challenge/", http.FileServer(http.Dir("/.well-known/acme-challenge"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL
		url.Scheme = "https"
		url.Host = r.Host
		http.Redirect(w, r, url.String(), http.StatusPermanentRedirect)
	})

	logger.Fatalln(http.ListenAndServe(":80", nil))
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

							expTime := time.UnixMilli(tokenInfo.Expires).Sub(time.Now())
							cacheTime := 3 * time.Minute
							
							if cacheTime > expTime {
								cacheTime = expTime
							}

							cacheStorage.SetCache("tokenInfo;"+authToken, map[string]string{
								"email":       email,
								"uid":         uid,
								"displayName": displayName,
							}, cacheTime)

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
