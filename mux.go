package mux

import (
	"net/http"
)

type Router struct {
	tree *node
}

func CreateRouter() *Router {
	return &Router{tree: createRootNode()}
}

func (r *Router) GET( path string, handler http.Handler) {
	r.tree.register("GET", path, handler)
}
func (r *Router) POST(path string, handler http.Handler) {
	r.tree.register("POST", path, handler)
}
func (r *Router) PUT(path string, handler http.Handler) {
	r.tree.register("PUT", path, handler)
}
func (r *Router) DELETE(path string, handler http.Handler) {
	r.tree.register("DELETE", path, handler)
}
func (r *Router) Handle(method string, path string, handler http.Handler) {
	r.tree.register(method, path, handler)
}

func (r *Router) HandleFunc(method string, path string, f func(http.ResponseWriter,
	*http.Request)) {
	r.Handle(method,path, nil)
}
