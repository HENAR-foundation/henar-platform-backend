package main

import (
	"henar-backend/db"
	"henar-backend/routes"
	"henar-backend/static"
	"log"
	"time"

	sentryfiber "github.com/aldy505/sentry-fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"

	_ "henar-backend/docs"

	"github.com/getsentry/sentry-go"
)

// @title Henar
// @version 1.0
// @host localhost:8080
// @BasePath /
func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              "https://ae6e801d7f7ef33cf287e5e8f306dc8a@o4506049431863296.ingest.sentry.io/4506049440972800",
		TracesSampleRate: 1.0,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	defer sentry.Flush(2 * time.Second)

	sentry.CaptureMessage("It works!")

	db.InitDb()

	static.Init()

	app := fiber.New()

	app.Use(logger.New())

	app.Use(sentryfiber.New(sentryfiber.Options{}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Use(cors.New(cors.Config{
		AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin,Access-Control-Allow-Credentials",
		AllowOrigins:     string("http://localhost:3000"),
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	routes.Setup(app)
}
