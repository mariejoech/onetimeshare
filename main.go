package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/django/v3"
	"log"
	"main.go/database/repository"
	"main.go/src"


	"github.com/joho/godotenv"
)


func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}
	err = repository.ConnectDB()
	if err != nil {
		fmt.Println("Could not connect to database")
		panic(err)
	}
	defer repository.DB.Close()

	// Create a new engine
	engine := django.New("./views", ".html")

	// Or from an embedded system
	// See github.com/gofiber/embed for examples
	// engine := html.NewFileSystem(http.Dir("./views", ".django"))

	// Pass the engine to the Views
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	src.SetupRoutes(app)

	//
	//app.Get("/embed", func(c *fiber.Ctx) error {
	//	// Render index within layouts/main
	//	return c.Render("embed", fiber.Map{
	//		"Title": "Hello, World!",
	//	}, "layouts/main2")
	//})

	log.Fatal(app.Listen(":3002"))
}
