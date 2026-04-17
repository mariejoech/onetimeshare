package src

import (
	"github.com/gofiber/fiber/v2"
	"main.go/src/controller"
	"main.go/src/middleware"
)

func SetupRoutes(app *fiber.App) {
	app.Use(middleware.HostValidation)

	pages := app.Group("/")

	pages.Get("/", controller.HomeGet)
	pages.Post("/", controller.HomePost)
	pages.Post("/burn", controller.BurnPost)
	pages.Get("/view", controller.ViewGet)
	pages.Post("/verify-password", controller.VerifyPassword)

	app.Use(controller.NotFoundPage)
}
