package main

import (
	"henar-backend/db"
	"henar-backend/routes"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"

	_ "henar-backend/docs"
)

// @title Henar
// @version 1.0
// @host localhost:8080
// @BasePath /
func main() {
	db.InitDb()

	log.Println("gsrg")
	app := fiber.New()

	app.Use(logger.New())

	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Use(cors.New(cors.Config{
		AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin,Access-Control-Allow-Credectials",
		AllowOrigins:     string("http://localhost:3000"),
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	routes.Setup(app)
}
