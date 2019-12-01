package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"net/http"
	"strings"
)

func HandleErr(err error) {
	//fmt.Printf("err")
	if err != nil {
		panic(err)
	}
}

// respondwithError return error message
func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondwithJSON(w, code, map[string]string{"message": msg})
}

// respondwithJSON write json response format
func respondwithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	fmt.Println(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

func UniformRand() float64 {
	rnd := rand.Float64()
	max := 1.0
	min := 0.0

	return rnd*(max-min) + min
}
