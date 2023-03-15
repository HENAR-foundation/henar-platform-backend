package main

import (
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

	router.HandleFunc("/v1/projects", projects.GetProjects).Methods("GET")
	router.HandleFunc("/v1/projects/{projectId}", projects.GetProject).Methods("GET")
	router.HandleFunc("/v1/projects", projects.CreateProject).Methods("POST")
	router.HandleFunc("/v1/projects/{projectId}", projects.UpdateProject).Methods("PATCH")
	router.HandleFunc("/v1/projects/{projectId}", projects.DeleteProject).Methods("DELETE")

	corsHandler := cors.AllowAll().Handler(router)
	http.ListenAndServe(":8080", corsHandler)
}
