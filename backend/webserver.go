package main

import (
	"bytes"
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
	"threadhelpServer/providers"
	"threadhelpServer/utils"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	midLogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/google/uuid"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var useOAuth = os.Getenv("USE_OAUTH") == "true"
var password = os.Getenv("PASSWORD")
var oauthAllowDomain = os.Getenv("OAUTH_ALLOW_DOMAIN")
var webpImageEncoding = os.Getenv("WEBP_IMAGE_ENCODING") == "true"
var httpsDomain = os.Getenv("HTTPS_DOMAIN")
var useHttps = os.Getenv("USE_HTTPS") == "true"
var sse = utils.NewSSEServer()

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
	var loginProvider providers.Provider
	if useOAuth {
		provider, err := providers.NewOAuthProvider(
			oauthAllowDomain,
			&cacheStorage,
		)
		if err != nil {
			return err
		}

		loginProvider = provider

		apiGroup.Get("check", func(c fiber.Ctx) error {
			if !loginProvider.CheckLogin(&c) {
				return c.Status(fiber.StatusUnauthorized).SendString(loginProvider.GetProviderName())
			}

			retInfo := map[string]any{}

			retInfo["admin"] = utils.IsAdmin(c.Locals("email").(string))

			return c.Status(fiber.StatusOK).JSON(retInfo)
		})
	} else {
		provider := providers.NewPasscodeProvider()
		loginProvider = provider

		apiGroup.Get("logout", func(c fiber.Ctx) error {
			if loginProvider.CheckLogin(&c) {
				c.Set("Set-Cookie", "Auth-Token=; expires=Thu, 01 Jan 1970 00:00:00 GMT;")
				return c.SendStatus(fiber.StatusOK)
			}

			return c.Status(fiber.StatusUnauthorized).SendString(loginProvider.GetProviderName())
		})

		apiGroup.Post("check", func(c fiber.Ctx) error {
			loggedIn := loginProvider.CheckLogin(&c)
			if loggedIn {
				user := providers.PasscodeUser{
					Name:     c.Locals("displayName").(string),
					Id:       c.Locals("uid").(string),
					IssuedAt: c.Locals("iat").(int64),
				}
				return c.Status(fiber.StatusOK).JSON(user)
			}

			var body map[string]string
			if json.Unmarshal(c.Body(), &body) != nil {
				return c.SendStatus(fiber.StatusBadRequest)
			}

			passcode := body["password"]
			if passcode == password {
				user := provider.GenerateNewUser()
				c.Cookie(&fiber.Cookie{
					Name:  "Auth-Token",
					Value: provider.GetUserToken(user),
				})
				return c.Status(fiber.StatusOK).JSON(user)
			}

			return c.Status(fiber.StatusUnauthorized).SendString(provider.GetProviderName())
		})
	}

	apiGroup.Get("provider", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString(loginProvider.GetProviderName())
	})

	apiGroup.Use(func(c fiber.Ctx) error {
		if !loginProvider.CheckLogin(&c) {
			return c.Status(fiber.StatusUnauthorized).SendString(loginProvider.GetProviderName())
		}

		return c.Next()
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

			for _, matched := range regexp.MustCompile("<\\s*([[:alpha:]]+)[ \\/>]").FindAllStringSubmatch(content, -1) {
				tagName := matched[1]
				if !slices.Contains(allowedTags, tagName) {
					return c.SendStatus(fiber.StatusBadRequest)
				}
			}

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

			clearExtraAttributes(doc.Nodes)

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

			post, err := utils.AddPost(utils.Post{
				UserID:          c.Locals("uid").(string),
				UserEmail:       c.Locals("email").(string),
				UserDisplayName: c.Locals("displayName").(string),
				Content:         content,
				AttachedImages:  attachedImages,
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

			sse.SendBytes(append([]byte("newPost;"), jsonData...))

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

		isAdmin := utils.IsAdmin(c.Locals("email").(string))

		var attachedImages []string
		var err error

		if isAdmin {
			attachedImages, err = utils.DeletePostAdmin(postId)
		} else {
			attachedImages, err = utils.DeletePost(postId, userId)
		}

		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		for _, img := range attachedImages {
			os.Remove("images/" + img)
		}

		sse.SendBytes([]byte("delPost;" + postId))

		return c.SendStatus(fiber.StatusOK)
	})

	apiGroup.Post("likePost", func(c fiber.Ctx) error {
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

		postUuid, err := uuid.Parse(postId)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusBadRequest)
		}

		err = utils.AddLike(userId, postUuid.String())
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		sse.SendBytes([]byte("updateLikes;" + postId))

		return c.SendStatus(fiber.StatusOK)
	})

	apiGroup.Post("unlikePost", func(c fiber.Ctx) error {
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

		postUuid, err := uuid.Parse(postId)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusBadRequest)
		}

		err = utils.RemoveLike(userId, postUuid.String())
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		sse.SendBytes([]byte("updateLikes;" + postId))

		return c.SendStatus(fiber.StatusOK)
	})

	apiGroup.Get("tenNewestPosts", func(c fiber.Ctx) error {
		posts, err := utils.GetNewestPosts(10)
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

		posts, err := utils.GetNewestPostsFrom(postId, 10)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.Status(fiber.StatusOK).JSON(posts)
	})

	apiGroup.Get("getPostContent/:postId", func(c fiber.Ctx) error {
		postId := c.Params("postId", "")
		if postId == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		postContent, err := utils.GetPostContent(postId)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).SendString("")
		}

		return c.Status(fiber.StatusOK).SendString(postContent)
	})

	apiGroup.Get("getPostLikes/:postId", func(c fiber.Ctx) error {
		postId := c.Params("postId", "")
		if postId == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		postLikes, err := utils.GetPostLikes(postId)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).SendString("{}")
		}

		liked, err := utils.CheckUserLikedPost(c.Locals("uid").(string), postId)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).SendString("{}")
		}

		return c.Status(fiber.StatusOK).JSON(map[string]any{
			"liked": liked,
			"likes": postLikes,
		})
	})

	{
		middlewaresSet := sse.FiberMiddlewaresSet()
		apiGroup.Get("/events", middlewaresSet[0], middlewaresSet[1:]...)
	}

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

func clearExtraAttributes(nodes []*html.Node) {
	for _, node := range nodes {
		if node.Type == html.ElementNode {
			{
				nodeName := node.DataAtom
				newAttributes := []html.Attribute{}

				for _, attr := range node.Attr {
					if (nodeName == atom.Img && (attr.Key == "src" || attr.Key == "alt")) ||
						(nodeName == atom.A && attr.Key == "href") {
						newAttributes = append(newAttributes, attr)
					}
				}

				node.Attr = newAttributes
			}

			children := []*html.Node{}
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				children = append(children, child)
			}
			clearExtraAttributes(children)
		}
	}
}
