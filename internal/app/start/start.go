package start

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func StartHTTPServer(port string) {
	r := mux.NewRouter()

	r.HandleFunc("/health", HealthCheck).Methods("GET")

	fmt.Println("Server is running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
