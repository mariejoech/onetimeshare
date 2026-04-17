package controller

import (
	"github.com/gofiber/fiber/v2"
	"log"
)

func NotFoundPage(c *fiber.Ctx) error {
	log.Print("page not found ", c.Hostname())
	return c.Status(fiber.StatusNotFound).Render("404", fiber.Map{})
}
