package mux

import "net/http"

type Group struct {
	Routers []*Router
}

func (g *Group) Handle(root string, router *Router) {
	var routes []*Route
	for _, route := range router.Routes {
		// skip the malformed route
		// with no root or patter given
		if root == "" || route.Pattern == "" {
			continue
		}
		// copying base to avoid
		// changes in root variable
		var base = root

		// fixing trailing/root slash
		// making sure no duplicates
		if base[len(base)-1:] == "/" && route.Pattern[0] == '/' {
			base = base[:len(base)-1]

			// adding slash if missing
		} else if base[len(base)-1:] != "/" && route.Pattern[0] != '/' {
			route.Pattern = "/" + route.Pattern
		}

		// internally converting asterisk
		// tokens to regex wildcard match
		if route.Pattern == "/*" {
			route.Pattern = "/.*"
			// if slash is not used then catch *
		} else if route.Pattern == "*" {
			route.Pattern = ".*"
		}

		// concatenation of the
		// group base and route pattern
		route.Pattern = base + route.Pattern
		routes = append(routes, route)
	}

	// re-adds the modified *routes
	// to the router
	router.Routes = routes
	// and appends *routers to group
	g.Routers = append(g.Routers, router)
}

func (g *Group) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, router := range g.Routers {
		if router.canServe(r) {
			router.ServeHTTP(w, r)
			return
		}
	}

	http.NotFound(w, r)
	return
}

func NewGroup(routers ...*Router) *Group {
	return &Group{
		Routers: routers,
	}
}
