package providers

import (
	"crypto/rand"
	"fmt"
	mrand "math/rand/v2"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type PasscodeUser struct {
	Name     string `json:"name"`
	Id       string `json:"id"`
	IssuedAt int64  `json:"iat"`
}

func (a PasscodeUser) Valid() error {
	if a.Name == "" {
		return fmt.Errorf("empty name field")
	}
	if err := uuid.Validate(a.Id); err != nil {
		return err
	}
	if time.Now().UnixMilli() < a.IssuedAt {
		return fmt.Errorf("invalid issuedAt field value")
	}

	return nil
}

type PasscodeProvider struct {
	secretKey string
}

func NewPasscodeProvider() PasscodeProvider {
	b := make([]byte, 32)
	rand.Read(b)

	return PasscodeProvider{
		secretKey: string(b),
	}
}

func (a *PasscodeProvider) GenerateNewUser() PasscodeUser {
	uid, err := uuid.NewUUID()

	for err != nil {
		uid, err = uuid.NewUUID()
	}

	namePreffixes := []string{"scary", "fast", "amber", "dark", "hollow", "soft", "little", "big", "great", "charming", "mystery", "blind", "wild", "busy", "awesome"}
	nameSuffixes := []string{"bear", "snake", "bee", "cowboy", "heart", "echo", "cat", "dog", "bird", "eagle", "fish", "boy", "girl", "dinosaur", "schoolboy", "schoolgirl", "ruby", "developer"}
	name := namePreffixes[mrand.IntN(len(namePreffixes))] + "-" + nameSuffixes[mrand.IntN(len(nameSuffixes))] + fmt.Sprint(mrand.Int32())

	user := PasscodeUser{
		Name:     name,
		Id:       uid.String(),
		IssuedAt: time.Now().UnixMilli(),
	}

	return user
}

func (a PasscodeProvider) GetUserToken(user PasscodeUser) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": user.Name,
		"id":   user.Id,
		"iat":  user.IssuedAt,
	})

	strToken, _ := token.SignedString([]byte(a.secretKey))

	return strToken
}

func (a PasscodeProvider) CheckLogin(c *fiber.Ctx) bool {
	strToken := (*c).Cookies("Auth-Token", "")

	if strToken == "" {
		return false
	}

	token, err := jwt.Parse(strToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(a.secretKey), nil
	}, jwt.WithoutClaimsValidation())
	if err != nil || !token.Valid {
		return false
	}
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	var nameStr string = ""
	var idStr string = ""
	var iatFloat float64 = 0.

	if nameClaim, ok := mapClaims["name"]; ok {
		if n, ok := nameClaim.(string); ok {
			nameStr = n
		} else {
			return false
		}
	} else {
		return false
	}

	if idClaim, ok := mapClaims["id"]; ok {
		if n, ok := idClaim.(string); ok {
			idStr = n
		} else {
			return false
		}
	} else {
		return false
	}

	if iatClaim, ok := mapClaims["iat"]; ok {
		if n, ok := iatClaim.(float64); ok {
			iatFloat = n
		} else {
			return false
		}
	} else {
		return false
	}

	claims := PasscodeUser{
		Name:     nameStr,
		Id:       idStr,
		IssuedAt: int64(iatFloat),
	}

	if claims.Valid() != nil {
		return false
	}

	(*c).Locals("displayName", claims.Name)
	(*c).Locals("uid", claims.Id)
	(*c).Locals("email", "-")

	(*c).Locals("iat", claims.IssuedAt)

	return true
}

func (a PasscodeProvider) GetProviderName() string {
	return "passcode"
}
