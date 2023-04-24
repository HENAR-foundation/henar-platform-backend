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

	authGroup := app.Group("/v1/auth")
	authGroup.Post("/signup", SignUp)
	authGroup.Post("/signin", SignIn)
	authGroup.Get("/signout", SignOut)
	authGroup.Get("/check", Check)

	// Locations routes
	locationsGroup := app.Group("/v1/locations")
	locationsGroup.Get("", locations.GetLocations)
	locationsGroup.Get("/suggestions", locations.GetLocationSuggestions)
	locationsGroup.Get("/:id", locations.GetLocation)

	locationsGroupSecured := app.Group("/v1/locations", SessionMiddleware)
	locationsGroupSecured.Post("", locations.CreateLocation)
	locationsGroupSecured.Patch("/:id", locations.UpdateLocation)
	locationsGroupSecured.Delete("/:id", locations.DeleteLocation)

	// Events routes
	eventsGroup := app.Group("/v1/events")
	eventsGroup.Get("", events.GetEvents)
	eventsGroup.Get("/:slug", events.GetEvent)

	eventsGroupSecured := app.Group("/v1/events", SessionMiddleware)
	eventsGroupSecured.Post("", events.CreateEvent)
	eventsGroupSecured.Patch("/:id", events.UpdateEvent)
	eventsGroupSecured.Delete("/:id", events.DeleteEvent)

	// Statistics routes
	statisticsGroup := app.Group("/v1/statistics")
	statisticsGroup.Get("", statistics.GetStatistics)
	statisticsGroup.Get("/:id", statistics.GetStatistic)

	statisticsGroupSecured := app.Group("/v1/statistics", SessionMiddleware)
	statisticsGroupSecured.Post("", statistics.CreateStatistic)
	statisticsGroupSecured.Patch("/:id", statistics.UpdateStatistic)
	statisticsGroupSecured.Delete("/:id", statistics.DeleteStatistic)

	// Tags routes
	tagsGroup := app.Group("/v1/tags")
	tagsGroup.Get("", tags.GetTags)
	tagsGroup.Get("/:id", tags.GetTag)

	tagsGroupSecured := app.Group("/v1/tags", SessionMiddleware)
	tagsGroupSecured.Post("", tags.CreateTag)
	tagsGroupSecured.Patch("/:id", tags.UpdateTag)
	tagsGroupSecured.Delete("/:id", tags.DeleteTag)

	// Projects routes
	projectsGroup := app.Group("/v1/projects")
	projectsGroup.Get("", projects.GetProjects)
	projectsGroup.Get("/:slug", projects.GetProject)

	projectsGroupSecured := app.Group("/v1/projects", SessionMiddleware)
	projectsGroupSecured.Get("/respond/:id", projects.RespondToProject(store))
	projectsGroupSecured.Post("", projects.CreateProject)
	projectsGroupSecured.Patch("/:id", projects.UpdateProject)
	projectsGroupSecured.Delete("/:id", projects.DeleteProject)

	// Researches routes
	researchesGroup := app.Group("/v1/researches")
	researchesGroup.Get("", researches.GetResearches)
	researchesGroup.Get("/:slug", researches.GetResearch)

	researchesGroupSecured := app.Group("/v1/researches", SessionMiddleware)
	researchesGroupSecured.Post("", researches.CreateResearch)
	researchesGroupSecured.Patch("/:id", researches.UpdateResearch)
	researchesGroupSecured.Delete("/:id", researches.DeleteResearch)

	app.Listen(":8080")
}
