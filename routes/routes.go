package routes

import (
	"henar-backend/events"
	"henar-backend/locations"
	"henar-backend/notifications"
	"henar-backend/projects"
	"henar-backend/researches"
	"henar-backend/static"
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

	authGroup := app.Group("/v1/auth")
	authGroup.Post("/signup", SignUp)
	authGroup.Post("/signin", SignIn)
	authGroup.Get("/signout", SignOut)
	authGroup.Get("/check", AuthorMiddleware, Check)
	authGroup.Post("/forgot-password", ForgotPassword)
	authGroup.Patch("/reset-password/:token", ResetPassword)

	// Locations routes
	locationsGroup := app.Group("/v1/locations")
	locationsGroup.Get("", locations.GetLocations)
	locationsGroup.Get("/suggestions", locations.GetLocationSuggestions)
	locationsGroup.Get("/:id", locations.GetLocation)
	locationsGroup.Post("", locations.CreateLocation)

	locationsGroupSecured := app.Group("/v1/locations", SessionMiddleware, AdminMiddleware)
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

	statisticsGroupSecured := app.Group("/v1/statistics", SessionMiddleware, AdminMiddleware)
	statisticsGroupSecured.Post("", statistics.CreateStatistic)
	statisticsGroupSecured.Patch("/:id", statistics.UpdateStatistic)
	statisticsGroupSecured.Delete("/:id", statistics.DeleteStatistic)

	// Tags routes
	tagsGroup := app.Group("/v1/tags")
	tagsGroup.Get("", tags.GetTags)
	tagsGroup.Get("/:id", tags.GetTag)

	tagsGroupSecured := app.Group("/v1/tags", SessionMiddleware, AdminMiddleware)
	tagsGroupSecured.Post("", tags.CreateTag)
	tagsGroupSecured.Patch("/:id", tags.UpdateTag)
	tagsGroupSecured.Delete("/:id", tags.DeleteTag)

	// Projects routes
	projectsGroup := app.Group("/v1/projects", AuthorMiddleware, AdminMiddleware)
	projectsGroup.Get("", projects.GetProjects)
	projectsGroup.Get("/:slug", projects.GetProject)

	projectsGroupSecured := app.Group("/v1/projects", SessionMiddleware, AdminMiddleware, AuthorMiddleware)
	projectsGroupSecured_TEMP := app.Group("/v1", SessionMiddleware, AdminMiddleware, AuthorMiddleware)
	projectsGroupSecured_TEMP.Get("/my-projects", projects.GetSelfProjects)
	projectsGroupSecured.Post("", projects.CreateProject)
	projectsGroupSecured.Post("/applicants/approve", projects.ApproveApplicant)
	projectsGroupSecured.Post("/applicants/reject", projects.RejectApplicant)
	projectsGroupSecured.Get("/respond/:id", projects.RespondToProject)
	projectsGroupSecured.Get("/cancel/:id", projects.CancelProjectApplication)
	// TODO: what if owner approve applicant?
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
	usersGroup := app.Group("/v1/users", AdminMiddleware, AuthorMiddleware)
	usersGroup.Get("", users.GetUsers)
	usersGroup.Get("/:id", users.GetUser)
	usersGroup.Post("", users.CreateUser)

	usersGroupSecured := app.Group("/v1/users", SessionMiddleware, AdminMiddleware, AuthorMiddleware)
	usersGroupSecured.Patch("/update-password", users.UpdatePassword)
	usersGroupSecured.Patch("/:id", users.UpdateUser)
	usersGroupSecured.Delete("/:id", users.DeleteUser)

	// user contacts request handlers
	usersGroupSecured.Post("request-contacts/:id", users.RequestContacts)
	usersGroupSecured.Get("approve-contacts-request/:id", users.ApproveContactsRequest)
	usersGroupSecured.Get("reject-contacts-request/:id", users.RejectContactsRequest)

	// user projects request handlers
	usersGroupSecured.Get("approve-project-request/:id", users.ApproveProjectRequest)
	usersGroupSecured.Get("reject-project-request/:id", users.RejectProjectRequest)

	staticGroup := app.Group("/v1/files")
	staticGroup.Post("/upload", static.UploadFile)

	notificationsGroupSecured := app.Group("/v1/notifications", SessionMiddleware, AdminMiddleware, AuthorMiddleware)
	notificationsGroupSecured.Get("", notifications.GetNotifications)
	notificationsGroupSecured.Post("", notifications.ReadNotifications)

	app.Listen(":8080")
}
