package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/gonum/stat/distuv"
	distuv2 "gonum.org/v1/gonum/stat/distuv"
	"math/rand"
	"net/http"
	"strings"
)

// HandleErr is a wrapper to panick if the error exists
func HandleErr(err error) {
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

// FileServer is the handler for serving files at a static route
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
func UniformRandMinMax(min float64, max float64) float64 {
	rnd := rand.Float64()
	return rnd*(max-min) + min
}

// UniformRand randomly picks something from 0 to 1
func UniformRand() float64 {
	max := 1.0
	min := 0.0
	return UniformRandMinMax(min, max)

}

// RightPad2Len pads a string to a certain length for better printing
func RightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}

func getExpRand(rate float64, cutoff float64, removeUnlikelyEvents bool) float64 {
	var exponential = distuv.Exponential{Rate: rate}

	movementTime := exponential.Rand()
	if !removeUnlikelyEvents {
		return movementTime
	}

	iterations := 0
	for !(exponential.Prob(movementTime) < cutoff) {
		movementTime = exponential.Rand()
		iterations += 1
		if iterations > unlikelyIterations {
			return exponential.Mean()
		}
	}

	return movementTime
}

func getPoissonRand(lambda float64, cutoff float64, removeUnlikelyEvents bool) (float64, float64) {
	var poisson = distuv2.Poisson{Lambda: lambda}

	movementTime := poisson.Rand()
	if !removeUnlikelyEvents {
		return movementTime, poisson.Prob(movementTime)
	}
	iterations := 0

	for !(poisson.Prob(movementTime) < cutoff) {
		movementTime = poisson.Rand()
		iterations += 1
		if iterations > unlikelyIterations {
			return poisson.Mean(), 0.5
		}
	}
	return movementTime, poisson.Prob(movementTime)
}
