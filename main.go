package main;

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"html/template"
	"net/http"
)

//var port int
//
//func renderTemplate(w http.ResponseWriter, tmpl string) {
//	t, err := template.ParseFiles("templates/" + tmpl)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	err = t.Execute(w, nil)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//	}
//}
//func handler(writer http.ResponseWriter, request *http.Request) {
//	renderTemplate(writer, "index.html")
//}
//
//func main() {
//	// Initialize the arguments when the main function is ran. This is to setup the settings needed by
//	// other parts of the file server.
//	flag.IntVar(&port, "p", 80, "Port to listen for HTTP requests (default port 8080).")
//	// Parse the args.
//	flag.Parse()
//	// Say that we are starting the server.
//	fmt.Printf("Server starting, port: http://localhost:%v\n", port)
//	serverString := fmt.Sprintf(":%v", port)
//
//	server := mux.NewRouter()
//	server.HandleFunc("/", handler)
//	//go RunSimulation()
//	fs := http.FileServer(http.Dir("/home/bob/static"))
//	http.Handle("/static/", http.StripPrefix("/static", fs))
//
//	log.Fatal(http.ListenAndServe(serverString, server))
//
//}
//package main

// TemplateRenderer is a custom html/template renderer for Echo framework
//type TemplateRenderer struct {
//	templates *template.Template
//}
//
//// Render renders a template document
//func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
//
//	// Add global methods if data is a map
//	if viewContext, isMap := data.(map[string]interface{}); isMap {
//		viewContext["reverse"] = c.Echo().Reverse
//	}
//
//	return t.templates.ExecuteTemplate(w, name, data)
//}
//
//func main() {
//	// Echo instance
//	e := echo.New()
//
//	// Middleware
//	e.Use(middleware.Logger())
//	e.Use(middleware.Recover())
//
//	renderer := &TemplateRenderer{
//		templates: template.Must(template.ParseGlob("*.html")),
//	}
//	e.Renderer = renderer
//
//	// Route => handler
//	e.GET("/", func(c echo.Context) error {
//		return c.String(http.StatusOK, "Hello, World!\n")
//	})
//
//	// Start server
//	e.Logger.Fatal(e.Start(":80"))
//}

var port int

func index(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("welcome"))
	t, err := template.ParseFiles("templates/index.html")
	catch(err)
	t.Execute(w, nil)
}

func addRoutes(router *chi.Mux) *chi.Mux {
	router.Get("/", index)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))

	return router
}

func main() {
	//flag.IntVar(&port, "p", 80, "Port to listen for HTTP requests (default port 8080).")
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
