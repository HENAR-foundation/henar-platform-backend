package projects

import "net/http"

func GetProject(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("get project test"))
}

func GetProjects(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("get projects test"))
}

func CreateProject(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("create projects test"))
}

func UpdateProject(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("update projects test"))
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("delete projects test"))
}
