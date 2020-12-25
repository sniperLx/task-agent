package api

import "net/http"

type localRoute struct {
	method  string
	path    string
	handler http.HandlerFunc
}

func (lr *localRoute) Method() string {
	return lr.method
}

func (lr *localRoute) Path() string {
	return lr.path
}

func (lr *localRoute) Handler() http.HandlerFunc {
	return lr.handler
}

func NewRoute(method, path string, handler http.HandlerFunc) Route {
	return &localRoute{method, path, handler}
}

func NewPostRoute(url string, handler http.HandlerFunc) Route {
	return NewRoute("POST", url, handler)
}

func NewGETRoute(url string, handler http.HandlerFunc) Route {
	return NewRoute("GET", url, handler)
}
