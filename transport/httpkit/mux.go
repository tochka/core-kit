package httpkit

import (
	"net/http"
	"regexp"

	"github.com/bmizerany/pat"
)

var ParamReplacer = regexp.MustCompile("{(?P<param_name>[\\w|-]+)[^}]*}")

type ServeMux interface {
	Add(meth string, pat string, h http.Handler)
	ServeHTTP(w http.ResponseWriter, r *http.Request)

	GetPathParam(req *http.Request, key string) string
}

func newPatRouter() *patRouter {
	return &patRouter{pat.New()}
}

type patRouter struct {
	router *pat.PatternServeMux
}

func (r *patRouter) Add(meth string, path string, h http.Handler) {
	r.router.Add(meth, ParamReplacer.ReplaceAllString(path, ":${param_name}"), h)
}

func (r *patRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func (r *patRouter) GetPathParam(req *http.Request, key string) string {
	return req.URL.Query().Get(":" + key)
}
