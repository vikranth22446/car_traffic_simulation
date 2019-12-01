package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	fps  = 24               // Frames per second.
	fpsl = 1000 / fps       // Duration of a single (milliseconds)
	fpsn = 1000000000 / fps // Duration of a single frame (nanoseconds)
)

var port int

func index(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("welcome"))
	t, err := template.ParseFiles("templates/index.html")
	HandleErr(err)

	t.Execute(w, nil)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Handler to /ws
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	addr := ws.RemoteAddr()

	log.Printf("Websocket accepted: %s\n", addr)

	user := newUser(ws)
	// New User is added to the main gas
	userGroup.addUser(user)

	go user.writer()

	err = user.identify()
	if err != nil {
		ws.Close()
		return
	}
	user.reader()

	// Once reader returns the connection is finalized
	log.Printf("Websocket finalized: %s\n", addr)
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func addRoutes(router *chi.Mux) *chi.Mux {
	workDir, err := os.Getwd()
	HandleErr(err)

	FileServer(router, "/static", http.Dir(filepath.Join(workDir, "public/build")))
	router.Get("/", index)
	router.HandleFunc("/ws", wsHandler)
	return router
}

func runServer() {
	flag.IntVar(&port, "p", 80, "Port to listen for HTTP requests (default port 8080).")
	// Parse the args.
	flag.Parse()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	addRoutes(r)

	fmt.Printf("Starting serve at http://localhost:%v\n", 80)
	http.ListenAndServe(":80", r)
}
