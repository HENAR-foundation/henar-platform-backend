package main

import (
	"henar-backend/db"
	"henar-backend/projects"
	"henar-backend/tags"
	"henar-backend/utils"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	db.InitDb()

	router := mux.NewRouter()

	router.Use(utils.RouterLoggerMiddleware)

	router.HandleFunc("/v1/tags", tags.GetTags).Methods("GET")
	router.HandleFunc("/v1/tags/{tagId}", tags.GetTag).Methods("GET")
	router.HandleFunc("/v1/projects", projects.GetProjects).Methods("GET")
	router.HandleFunc("/v1/projects/{projectId}", projects.GetProject).Methods("GET")
	router.HandleFunc("/v1/projects", projects.CreateProject).Methods("POST")
	router.HandleFunc("/v1/projects/{projectId}", projects.UpdateProject).Methods("PATCH")
	router.HandleFunc("/v1/projects/{projectId}", projects.DeleteProject).Methods("DELETE")

	corsHandler := cors.AllowAll().Handler(router)
	http.ListenAndServe(":8080", corsHandler)
}
