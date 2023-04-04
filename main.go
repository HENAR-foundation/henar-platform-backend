package main

import (
	"henar-backend/db"
	"henar-backend/events"
	"henar-backend/projects"
	"henar-backend/researches"
	"henar-backend/statistics"
	"henar-backend/tags"

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

	app := fiber.New()

	app.Use(logger.New())

	app.Get("/swagger/*", swagger.HandlerDefault)

	// Events routes
	eventsGroup := app.Group("/v1/events")
	eventsGroup.Get("", events.GetEvents)
	eventsGroup.Get("/:id", events.GetEvent)
	eventsGroup.Post("", events.CreateEvent)
	eventsGroup.Patch("/:id", events.UpdateEvent)
	eventsGroup.Delete("/:id", events.DeleteEvent)

	// Statistics routes
	statisticsGroup := app.Group("/v1/statistics")
	statisticsGroup.Get("", statistics.GetStatistics)
	statisticsGroup.Get("/:id", statistics.GetStatistic)
	statisticsGroup.Post("", statistics.CreateStatistic)
	statisticsGroup.Patch("/:id", statistics.UpdateStatistic)
	statisticsGroup.Delete("/:id", statistics.DeleteStatistic)

	// Tags routes
	tagsGroup := app.Group("/v1/tags")
	tagsGroup.Get("", tags.GetTags)
	tagsGroup.Get("/:id", tags.GetTag)
	tagsGroup.Post("", tags.CreateTag)
	tagsGroup.Patch("/:id", tags.UpdateTag)
	tagsGroup.Delete("/:id", tags.DeleteTag)

	// Projects routes
	projectsGroup := app.Group("/v1/projects")
	projectsGroup.Get("", projects.GetProjects)
	projectsGroup.Get("/:id", projects.GetProject)
	projectsGroup.Post("", projects.CreateProject)
	projectsGroup.Patch("/:id", projects.UpdateProject)
	projectsGroup.Delete("/:id", projects.DeleteProject)

	// Researches routes
	researchesGroup := app.Group("/v1/researches")
	researchesGroup.Get("", researches.GetResearches)
	researchesGroup.Get("/:id", researches.GetResearch)
	researchesGroup.Post("", researches.CreateResearch)
	researchesGroup.Patch("/:id", researches.UpdateResearch)
	researchesGroup.Delete("/:id", researches.DeleteResearch)

	app.Use(cors.New())

	app.Listen(":8080")
}
