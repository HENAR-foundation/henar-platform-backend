package main

import (
	"henar-backend/db"
	"henar-backend/routes"
	"henar-backend/static"
	"log"
	"os"
	"time"

	sentryfiber "github.com/aldy505/sentry-fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"

	_ "henar-backend/docs"

	"github.com/getsentry/sentry-go"
)

func init() {
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load(".env")

		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
}

// @title Henar
// @version 1.0
// @host localhost:8080
// @BasePath /
func main() {
	sentryDsn := os.Getenv("SENTRY_DSN")

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDsn,
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
