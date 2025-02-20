package providers

import (
	"context"
	"strings"
	"threadhelpServer/utils"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v3"
	"google.golang.org/api/option"
)

type OAuthProvider struct {
	firebaseAuth *auth.Client
	cacheStorage *utils.CacheStorage
	AllowDomain  string
}

func NewOAuthProvider(allowDomain string, cacheStorage *utils.CacheStorage) (OAuthProvider, error) {
	opt := option.WithCredentialsFile("./firebaseSecretKey.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return OAuthProvider{}, err
	}

	firebaseAuth, err := app.Auth(context.Background())
	if err != nil {
		return OAuthProvider{}, err
	}

	return OAuthProvider{
		firebaseAuth: firebaseAuth,
		AllowDomain:  allowDomain,
		cacheStorage: cacheStorage,
	}, nil
}

func (a OAuthProvider) CheckLogin(c *fiber.Ctx) bool {
	var email, uid, displayName string

	if !func() bool {
		authToken := (*c).Get("Auth-Token", "")
		if len(authToken) == 0 {
			return false
		}

		if tokenInfo, ok := a.cacheStorage.GetCache("tokenInfo;" + authToken); ok {
			if mapTokenInfo, ok := tokenInfo.(map[string]string); ok {
				email = mapTokenInfo["email"]
				uid = mapTokenInfo["uid"]
				displayName = mapTokenInfo["displayName"]

				return true
			}
		}

		tokenInfo, err := a.firebaseAuth.VerifyIDTokenAndCheckRevoked(context.Background(), authToken)
		if err != nil {
			return false
		}

		if aEmail, ok := tokenInfo.Claims["email"]; ok {
			if strEmail, ok := aEmail.(string); ok {
				if strings.HasSuffix(strEmail, "@"+a.AllowDomain) {
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

							a.cacheStorage.SetCache("tokenInfo;"+authToken, map[string]string{
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
		return false
	}

	(*c).Locals("email", email)
	(*c).Locals("uid", uid)
	(*c).Locals("displayName", displayName)

	if utils.IsInBlacklist(email) {
		return false
	}
	return true
}

func (a OAuthProvider) GetProviderName() string {
	return "oauth"
}
