package andrew

import (
	"fmt"
	"net/http"
)

func ServeRoot(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "ABC")
}
