package api

import "net/http"

type Router interface {
	Routes() []Route
}

type Route interface {
	Method() string
	Path() string
	Handler() http.HandlerFunc
}
