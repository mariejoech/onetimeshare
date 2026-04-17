package middleware

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"os"
)

func HostValidation(c *fiber.Ctx) error {

	fmt.Println("Middleware Host Validation")

	fmt.Println(os.Getenv("HOST_NAME"))
	fmt.Println(c.Hostname())

	if os.Getenv("HOST_NAME") != c.Hostname() {

		fmt.Println("if statement")
		return c.Status(fiber.StatusBadRequest).Render("errors/bad_request", fiber.Map{
			"Message":  "you are using invalid host name",
			"HostName": os.Getenv("HOST_NAME"),
		})

	}

	return c.Next()
}
