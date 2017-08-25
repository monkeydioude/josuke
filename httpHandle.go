package josuke

import (
	"fmt"
	"net/http"
)

func HttpHandle(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprint(rw, "ohayooo")
}
