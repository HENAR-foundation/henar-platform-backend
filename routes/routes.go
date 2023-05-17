package routes

import (
	"henar-backend/events"
	"henar-backend/locations"
	"henar-backend/projects"
	"henar-backend/researches"
	"henar-backend/statistics"
	"henar-backend/tags"
	"henar-backend/users"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var (
	store     *session.Store
	AUTH_KEY  string = "authentificated"
	USER_ID   string = "user_id"
	USER_ROLE string = "user_role"
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
	eventsGroup := app.Group("/v1/events", AdminMiddleware, AuthorMiddleware)
	eventsGroup.Get("", events.GetEvents)
	eventsGroup.Get("/:slug", events.GetEvent)

	eventsGroupSecured := app.Group("/v1/events", SessionMiddleware, AdminMiddleware, AuthorMiddleware)
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
	projectsGroup := app.Group("/v1/projects", AuthorMiddleware, AdminMiddleware)
	projectsGroup.Get("", projects.GetProjects)
	projectsGroup.Get("/:slug", projects.GetProject)

	projectsGroupSecured := app.Group("/v1/projects", SessionMiddleware, AdminMiddleware, AuthorMiddleware)
	projectsGroupSecured.Post("", projects.CreateProject(store))
	projectsGroupSecured.Get("/respond/:id", projects.RespondToProject(store))
	projectsGroupSecured.Patch("/:id", projects.UpdateProject)
	projectsGroupSecured.Delete("/:id", projects.DeleteProject(store))

	// Researches routes
	researchesGroup := app.Group("/v1/researches", AdminMiddleware, AuthorMiddleware)
	researchesGroup.Get("", researches.GetResearches)
	researchesGroup.Get("/:slug", researches.GetResearch)

	researchesGroupSecured := app.Group("/v1/researches", SessionMiddleware, AdminMiddleware, AuthorMiddleware)
	researchesGroupSecured.Post("", researches.CreateResearch)
	researchesGroupSecured.Patch("/:id", researches.UpdateResearch)
	researchesGroupSecured.Delete("/:id", researches.DeleteResearch)

	// User routes
	usersGroup := app.Group("/v1/users")
	usersGroup.Get("", users.GetUsers)
	usersGroup.Get("/:id", users.GetUser)

	// usersGroupSecured := app.Group("/v1/users", SessionMiddleware)
	usersGroup.Post("", users.CreateUser)
	usersGroup.Patch("/:id", users.UpdateUser)
	usersGroup.Delete("/:id", users.DeleteUser)

	app.Listen(":8080")
}
