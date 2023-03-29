package main

import (
	"henar-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin,Access-Control-Allow-Credectials",
		AllowOrigins:     string("http://localhost:3000"),
		AllowCredentials: true,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	routes.Setup(app)
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
