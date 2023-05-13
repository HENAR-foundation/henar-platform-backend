package routes

import (
	"henar-backend/events"
	"henar-backend/locations"
	"henar-backend/notifications"
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

	// locationsGroupSecured := app.Group("/v1/locations", SessionMiddleware)
	locationsGroup.Post("", locations.CreateLocation)
	locationsGroup.Patch("/:id", locations.UpdateLocation)
	locationsGroup.Delete("/:id", locations.DeleteLocation)

	// Events routes
	eventsGroup := app.Group("/v1/events")
	eventsGroup.Get("", events.GetEvents)
	eventsGroup.Get("/:slug", events.GetEvent)

	// 	eventsGroupSecured := app.Group("/v1/events", SessionMiddleware)
	eventsGroup.Post("", events.CreateEvent)
	eventsGroup.Patch("/:id", events.UpdateEvent)
	eventsGroup.Delete("/:id", events.DeleteEvent)

	// Statistics routes
	statisticsGroup := app.Group("/v1/statistics")
	statisticsGroup.Get("", statistics.GetStatistics)
	statisticsGroup.Get("/:id", statistics.GetStatistic)

	// statisticsGroupSecured := app.Group("/v1/statistics", SessionMiddleware)
	statisticsGroup.Post("", statistics.CreateStatistic)
	statisticsGroup.Patch("/:id", statistics.UpdateStatistic)
	statisticsGroup.Delete("/:id", statistics.DeleteStatistic)

	// Tags routes
	tagsGroup := app.Group("/v1/tags")
	tagsGroup.Get("", tags.GetTags)
	tagsGroup.Get("/:id", tags.GetTag)

	// tagsGroupSecured := app.Group("/v1/tags", SessionMiddleware)
	tagsGroup.Post("", tags.CreateTag)
	tagsGroup.Patch("/:id", tags.UpdateTag)
	tagsGroup.Delete("/:id", tags.DeleteTag)

	// Projects routes
	projectsGroup := app.Group("/v1/projects")
	projectsGroup.Get("", projects.GetProjects)
	projectsGroup.Get("/:slug", projects.GetProject)

	// projectsGroupSecured := app.Group("/v1/projects", SessionMiddleware)
	projectsGroup.Get("/respond/:id", projects.RespondToProject(store))
	projectsGroup.Post("", projects.CreateProject)
	projectsGroup.Patch("/:id", projects.UpdateProject)
	projectsGroup.Delete("/:id", projects.DeleteProject)

	// Researches routes
	researchesGroup := app.Group("/v1/researches")
	researchesGroup.Get("", researches.GetResearches)
	researchesGroup.Get("/:slug", researches.GetResearch)

	// 	researchesGroupSecured := app.Group("/v1/researches", SessionMiddleware)
	researchesGroup.Post("", researches.CreateResearch)
	researchesGroup.Patch("/:id", researches.UpdateResearch)
	researchesGroup.Delete("/:id", researches.DeleteResearch)

	// User routes
	usersGroup := app.Group("/v1/users")
	usersGroup.Get("", users.GetUsers)
	usersGroup.Get("/:id", users.GetUser)

	// usersGroupSecured := app.Group("/v1/users", SessionMiddleware)
	usersGroup.Post("", users.CreateUser)
	usersGroup.Patch("/:id", users.UpdateUser)
	usersGroup.Delete("/:id", users.DeleteUser)

	notificationsGroupSecured := app.Group("/v1/notifications", SessionMiddleware)
	notificationsGroupSecured.Get("", notifications.GetNotifications)
	notificationsGroupSecured.Post("/read", notifications.ReadNotifications)

	app.Listen(":8080")
}
