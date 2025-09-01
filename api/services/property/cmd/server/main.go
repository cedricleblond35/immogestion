package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Endpoint de healthcheck
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	port := "8082" // ou récupéré depuis os.Getenv("PORT")
	log.Printf("Auth service started on port %s", port)

	// Démarrer le serveur
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Erreur au démarrage du serveur: %v", err)
	}

}
