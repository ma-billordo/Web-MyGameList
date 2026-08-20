package main

import (
	"fmt"
	"net/http"
)

////Main////

func main() {
	http.HandleFunc("/", serveForm)
	http.HandleFunc("/about.html", handleAbout)

	port := ":8080"
	fmt.Printf("Servidor escuchando en http://localhost%s\n", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Printf("Error al iniciar el servidor: %s\n", err)
	}
}

////Handlers////

func serveForm(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "index.html")
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "about.html")
}

////Funciones////
