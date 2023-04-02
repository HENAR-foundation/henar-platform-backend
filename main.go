package main

import (
	"henar-backend/db"
	"henar-backend/projects"
	"henar-backend/researches"
	"henar-backend/tags"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	db.InitDb()

	app := fiber.New()

	app.Use(logger.New())

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
