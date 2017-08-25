package josuke

import (
	"fmt"
	"net/http"
)

type Handler struct {
	Uri string
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.RequestURI != fmt.Sprintf("/%s", h.Uri) {
		return
	}
	Request(rw, r)
}
