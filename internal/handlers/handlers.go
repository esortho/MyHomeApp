package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"runtime"
)

var templates = func() *template.Template {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../..")
	templatePath := filepath.Join(projectRoot, "web/templates/*.html")
	tmpl, err := template.ParseGlob(templatePath)
	if err != nil {
		panic(err)
	}
	return tmpl
}()

// IndexHandler handles requests to the index page
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// MeasurementsHandler handles requests to the measurements page
func MeasurementsHandler(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "measurements.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// LightsHandler handles requests to the lights page
func LightsHandler(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "lights.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// DashboardHandler handles requests to the dashboard page
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "dashboard.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Dashboard2Handler handles requests to the dashboard2 page
func Dashboard2Handler(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "dashboard2.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HistoryHandler handles requests to the history page
func HistoryHandler(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "history.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// PoolHandler handles requests to the pool page
func PoolHandler(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "pool.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
