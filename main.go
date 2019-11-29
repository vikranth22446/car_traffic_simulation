package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"html/template"
	"net/http"
)

var port int

func index(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("welcome"))
	t, err := template.ParseFiles("templates/index.html")
	Catch(err)

	t.Execute(w, nil)
}

func addRoutes(router *chi.Mux) *chi.Mux {
	router.Get("/", index)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))

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
