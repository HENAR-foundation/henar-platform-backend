package routes

import (
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
	// router.Use(BaseMiddleware, cors.New(cors.Config{
	// 	AllowCredentials: true,
	// 	AllowOrigins:     "*",
	// 	AllowHeaders:     "Access-Control-Allow-Origin, Content-Type, Authorization, Origin, Accept",
	// }))

	// noAuth := router.Group("")

	app.Post("/auth/signup", SignUp)
	app.Post("/auth/signin", SignIn)
	app.Get("/auth/signout", SignOut)
	app.Get("/auth/check", Check)

	app.Listen(":8080")
	// noAuth.Post("/auth/check", auth.Check)
	// router := mux.NewRouter()
	// router.Use(utils.RouterLoggerMiddleware)

	// corsHandler := cors.Default().Handler(router)

	// apiRouter := router.PathPrefix("/api").Subrouter()

	// noAuth := apiRouter.PathPrefix("").Subrouter()
	// {
	// 	noAuth.HandleFunc(("/auth/signup"), users.CreateUser).Methods("POST")
	// 	noAuth.HandleFunc("/auth/login", auth.LoginHandler).Methods("POST")
	// 	noAuth.HandleFunc("/auth/logout", auth.LogoutHandler).Methods("GET")
	// 	// noAuth.HandleFunc("/auth/check", auth.CheckAuthHandler).Methods("GET")
	// 	noAuth.HandleFunc("/projects/{projectId}", projects.GetProject).Methods("GET")
	// 	noAuth.HandleFunc("/projects", projects.GetProjects).Methods("GET")
	// }

	// sessionAuth := apiRouter.PathPrefix("").Subrouter()
	// sessionAuth.Use(auth.SessionMiddleware)
	// {
	// 	sessionAuth.HandleFunc("/users/{userId}", users.GetUser).Methods("GET")
	// 	sessionAuth.HandleFunc("/projects", projects.CreateProject).Methods("POST")
	// 	sessionAuth.HandleFunc("/projects/{projectId}", projects.UpdateProject).Methods("PATCH")
	// 	sessionAuth.HandleFunc("/projects/{projectId}", projects.DeleteProject).Methods("DELETE")
	// }

	// adminAuth := sessionAuth.PathPrefix("").Subrouter()
	// adminAuth.Use(auth.AdminMiddleware)
	// {

	// }

	// http.ListenAndServe(":8080", corsHandler)
}
