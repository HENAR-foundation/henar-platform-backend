package main

import (
	"henar-backend/auth"
	"henar-backend/projects"
	"henar-backend/utils"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	router := mux.NewRouter()
	router.Use(utils.RouterLoggerMiddleware)

	apiRouter := router.PathPrefix("/api").Subrouter()

	noAuth := apiRouter.PathPrefix("").Subrouter()
	{
		noAuth.HandleFunc("/login", auth.LoginHandler).Methods("POST")
		noAuth.HandleFunc("/logout", auth.LogoutHandler).Methods("GET")
		noAuth.HandleFunc("/projects/{projectId}", projects.GetProject).Methods("GET")
		noAuth.HandleFunc("/projects", projects.GetProjects).Methods("GET")
	}

	sessionAuth := apiRouter.PathPrefix("").Subrouter()
	sessionAuth.Use(auth.SessionMiddleware)
	{
		sessionAuth.HandleFunc("/projects", projects.CreateProject).Methods("POST")
		sessionAuth.HandleFunc("/projects/{projectId}", projects.UpdateProject).Methods("PATCH")
		sessionAuth.HandleFunc("/projects/{projectId}", projects.DeleteProject).Methods("DELETE")
	}

	adminAuth := sessionAuth.PathPrefix("").Subrouter()
	adminAuth.Use(auth.AdminMiddleware)
	{

	}

	corsHandler := cors.AllowAll().Handler(router)
	http.ListenAndServe(":8080", corsHandler)
}
