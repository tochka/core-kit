package corekit

import (
	"net/http"

	"github.com/bmizerany/pat"
)

type adoptPatRouter struct {
	router *pat.PatternServeMux
}

func (r *adoptPatRouter) Add(meth string, path string, h http.Handler) {
	r.router.Add(meth, path, h)
}

func (r *adoptPatRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}
