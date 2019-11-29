package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

var port int

func index(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("welcome"))
	t, err := template.ParseFiles("templates/index.html")
	HandleErr(err)

	t.Execute(w, nil)
}

func addRoutes(router *chi.Mux) *chi.Mux {
	workDir, err := os.Getwd()
	HandleErr(err)

	FileServer(router, "/static", http.Dir(filepath.Join(workDir, "static")))
	router.Get("/", index)

	return router
}

func main() {
	flag.IntVar(&port, "p", 80, "Port to listen for HTTP requests (default port 8080).")
	// Parse the args.
	flag.Parse()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	//r.Get("/", )
	addRoutes(r)
	fmt.Printf("Starting serve at http://localhost:%v\n", 80)
	http.ListenAndServe(":80", r)
}
