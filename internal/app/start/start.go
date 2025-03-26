package start

import (
	"fmt"
	"net/http"
)

func StartHTTPServer(port string, handler http.Handler) {
	fmt.Println("Server is running on port", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		panic(err)
	}
}
