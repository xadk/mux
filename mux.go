package mux

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type ctxKey struct{}

func parseRouteParams(r *http.Request, pattern string) (url.Values, error) {
	// PCRE(2) dirty:
	// `(?(?=\{)(?:(?:\{[\s]*)([^\:\/\{\}]+?)(?:\s*?\:\s*?))?(\(.*?\))(?:[\s]*\})|(?:([^\:\/\{\}]+?)(?:\s*?\:\s*?))?(\(.*?\)))`gm
	var re = regexp.MustCompile(`(?m)(?:\{[\s]*?(?:([^:/{}\s]+?)(?:\s*?\:\s*?))?(\(.*?\))(?:[\s]*?\})|(?:([^:/{}\s]+?)(?:[\s]*?\:[\s]*?))?(\(.*?\)))`)

	if pattern == "" {
		var ctxValue = r.Context().Value(ctxKey{})
		// and if no pre ctx
		// then simply return
		if ctxValue == nil {
			return url.Values{},
				fmt.Errorf("no routing pattern in ctx")
		}
		// pattern from ctx value
		pattern = ctxValue.(string)
	}

	// accept route if contains
	// a wildcard match pattern
	if pattern == "/" ||
		pattern == "*" || pattern == "/*" {
		return url.Values{}, nil
	}

	var keys []string
	for _, match := range re.FindAllStringSubmatch(pattern, -1) {
		var key, val string
		key, val = match[1], match[2]
		// non-bracket matches
		// or operator as m[3]
		// might be keyless
		if match[3] != "" || match[4] != "" {
			key = match[3]
			val = match[4]
		}
		// unnamed params
		if key == "" {
			key = "_"
		}
		// pattern with user rgx only
		// so can be executed as rgx
		pattern = strings.ReplaceAll(
			pattern,
			match[0],
			val,
		)
		keys = append(keys, key)
	}

	var dict = url.Values{}
	var matches = regexp.MustCompile(
		fmt.Sprintf(`(?m)^%s$`, pattern),
	).FindAllStringSubmatch(r.URL.Path, -1)

	if len(matches) < 1 {
		return dict,
			fmt.Errorf("url did not match the routing pattern")
	}

	for _, subMatches := range matches {
		if len(subMatches) < 1 {
			return dict,
				fmt.Errorf("routing pattern did not have any submatches")
		}

		for i, m := range subMatches[1:] {
			// ignore the overflown subMatches
			// in case of invalid regex
			// i.e. multiple brackets or
			// if doesn't matched by above regex
			if i >= len(keys) {
				break
			}
			dict.Add(keys[i], m)
		}
	}

	return dict, nil
}

func isValidRoute(r *http.Request, pattern string) bool {
	_, err := parseRouteParams(r, pattern)
	return err == nil
}

func Vars(r *http.Request) url.Values {
	dict, _ := parseRouteParams(r, "")
	return dict
}

func UnnamedVars(r *http.Request) []string {
	var vars = Vars(r)
	if !vars.Has("_") {
		return []string{}
	}
	return vars["_"]
}
