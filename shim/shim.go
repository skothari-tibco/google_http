package shim

import (
	"net/http"

	fl "github.com/project-flogo/google_http"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	fl.Invoke(w, r)
}
