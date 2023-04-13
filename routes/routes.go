package routes

import (
	"henar-backend/events"
	"henar-backend/locations"
	"henar-backend/projects"
	"henar-backend/researches"
	"henar-backend/statistics"
	"henar-backend/tags"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var (
	store    *session.Store
	AUTH_KEY string = "authentificated"
	USER_ID  string = "user_id"
)

func Setup(app *fiber.App) {
	store = session.New(session.Config{
		CookieHTTPOnly: true,
		Expiration:     time.Hour * 3000,
	})

	app.Post("/auth/signup", SignUp)
	app.Post("/auth/signin", SignIn)
	app.Get("/auth/signout", SignOut)
	app.Get("/auth/check", Check)

	locationsGroup := app.Group("/v1/locations")
	locationsGroup.Get("", locations.GetLocations)
	locationsGroup.Get("/suggestions", locations.GetLocationSuggestions)
	locationsGroup.Get("/:id", locations.GetLocation)
	locationsGroup.Post("", locations.CreateLocation)
	locationsGroup.Patch("/:id", locations.UpdateLocation)
	locationsGroup.Delete("/:id", locations.DeleteLocation)

	// Events routes
	eventsGroup := app.Group("/v1/events")
	eventsGroup.Get("", events.GetEvents)
	eventsGroup.Get("/:slug", events.GetEvent)
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
	projectsGroup.Get("/:slug", projects.GetProject)
	projectsGroup.Post("", projects.CreateProject)
	projectsGroup.Patch("/:id", projects.UpdateProject)
	projectsGroup.Delete("/:id", projects.DeleteProject)

	// Researches routes
	researchesGroup := app.Group("/v1/researches")
	researchesGroup.Get("", researches.GetResearches)
	researchesGroup.Get("/:slug", researches.GetResearch)
	researchesGroup.Post("", researches.CreateResearch)
	researchesGroup.Patch("/:id", researches.UpdateResearch)
	researchesGroup.Delete("/:id", researches.DeleteResearch)

	app.Listen(":8080")
}
