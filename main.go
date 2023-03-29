package main

import (
	"henar-backend/auth"
	"henar-backend/projects"
	"henar-backend/users"
	"henar-backend/utils"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// auth.InitSessions()
	router := mux.NewRouter()
	router.Use(utils.RouterLoggerMiddleware)

	corsHandler := cors.Default().Handler(router)

	apiRouter := router.PathPrefix("/api").Subrouter()

	noAuth := apiRouter.PathPrefix("").Subrouter()
	{
		noAuth.HandleFunc(("/auth/signup"), users.CreateUser).Methods("POST")
		noAuth.HandleFunc("/auth/login", auth.LoginHandler).Methods("POST")
		noAuth.HandleFunc("/auth/logout", auth.LogoutHandler).Methods("GET")
		// noAuth.HandleFunc("/auth/check", auth.CheckAuthHandler).Methods("GET")
		noAuth.HandleFunc("/projects/{projectId}", projects.GetProject).Methods("GET")
		noAuth.HandleFunc("/projects", projects.GetProjects).Methods("GET")
	}

	sessionAuth := apiRouter.PathPrefix("").Subrouter()
	sessionAuth.Use(auth.SessionMiddleware)
	{
		sessionAuth.HandleFunc("/users/{userId}", users.GetUser).Methods("GET")
		sessionAuth.HandleFunc("/projects", projects.CreateProject).Methods("POST")
		sessionAuth.HandleFunc("/projects/{projectId}", projects.UpdateProject).Methods("PATCH")
		sessionAuth.HandleFunc("/projects/{projectId}", projects.DeleteProject).Methods("DELETE")
	}

	adminAuth := sessionAuth.PathPrefix("").Subrouter()
	adminAuth.Use(auth.AdminMiddleware)
	{

	}

	http.ListenAndServe(":8080", corsHandler)
}
