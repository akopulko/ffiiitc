package router

import (
	"fmt"
	"log"
	"net/http"
)

type Router struct {
	Mux *http.ServeMux
}

func NewRouter() *Router {
	return &Router{
		Mux: http.NewServeMux(),
	}
}

func (r *Router) AddRoute(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	r.Mux.HandleFunc(pattern, handler)
}

func (r *Router) logRoute(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func (r *Router) Run(port int) error {
	ps := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(ps, r.logRoute(r.Mux))
}
