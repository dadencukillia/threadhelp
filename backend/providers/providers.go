package providers

import "github.com/gofiber/fiber/v3"

type Provider interface {
	GetProviderName() string
	CheckLogin(*fiber.Ctx) bool
}
