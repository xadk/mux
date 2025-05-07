package mux

import (
	"context"
	"net/http"
	"strings"
)

type Route struct {
	Method  string
	Pattern string
	Handler http.HandlerFunc
}

type Router struct {
	Routes []*Route
}

func (router *Router) Add(method, pattern string, handler http.HandlerFunc) {
	router.Routes = append(router.Routes, &Route{
		Method:  method,
		Pattern: pattern,
		Handler: handler,
	})
}

func (router *Router) All(pattern string, handler http.HandlerFunc) {
	router.Add("", pattern, handler)
}

func (router *Router) Get(pattern string, handler http.HandlerFunc) {
	router.Add("GET", pattern, handler)
}

func (router *Router) Head(pattern string, handler http.HandlerFunc) {
	router.Add("HEAD", pattern, handler)
}

func (router *Router) Post(pattern string, handler http.HandlerFunc) {
	router.Add("POST", pattern, handler)
}

func (router *Router) Put(pattern string, handler http.HandlerFunc) {
	router.Add("PUT", pattern, handler)
}

func (router *Router) Delete(pattern string, handler http.HandlerFunc) {
	router.Add("DELETE", pattern, handler)
}

func (router *Router) Connect(pattern string, handler http.HandlerFunc) {
	router.Add("CONNECT", pattern, handler)
}

func (router *Router) Options(pattern string, handler http.HandlerFunc) {
	router.Add("OPTIONS", pattern, handler)
}

func (router *Router) Trace(pattern string, handler http.HandlerFunc) {
	router.Add("TRACE", pattern, handler)
}

func (router *Router) Patch(pattern string, handler http.HandlerFunc) {
	router.Add("PATCH", pattern, handler)
}

func (router *Router) addCtx(r *http.Request, route *Route) {
	var ctxValue = r.Context().Value(ctxKey{})
	if ctxValue != nil {
		// ctx value is kept immutable
		if ctxValue.(string) != "" {
			return
		}
	}
	ctx := context.WithValue(r.Context(), ctxKey{}, route.Pattern)
	*r = *r.WithContext(ctx)
}

func (router *Router) canServe(r *http.Request) bool {
	for _, route := range router.Routes {
		if isValidRoute(r, route.Pattern) {
			return true
		}
	}
	return false
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var allowedMethods []string
	for _, route := range router.Routes {
		if isValidRoute(r, route.Pattern) {
			// if no route.Method is specified then serve
			if route.Method != "" && r.Method != route.Method {
				allowedMethods = append(allowedMethods, route.Method)
				continue
			}
			// stamping ctx before
			// handing over to handler
			router.addCtx(r, route)
			route.Handler(w, r)
			return
		}
	}

	if len(allowedMethods) > 0 {
		w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.NotFound(w, r)
	return
}

func NewRouter(routes ...*Route) *Router {
	return &Router{
		Routes: routes,
	}
}
