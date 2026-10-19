package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	db "MyGameList.com/db/sqlc"
	"MyGameList.com/internal/middleware"
	"MyGameList.com/pkg/handlers"

	_ "github.com/lib/pq"
)

//// Funciones /////

func abrirDB() (*sql.DB, error) {
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		return nil, fmt.Errorf("la variable de entorno DB_URL no está seteada")
	}
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}
	conn.SetMaxOpenConns(25)
	return conn, nil
}

//// Handlers ////

func serveIndex(w http.ResponseWriter, r *http.Request) {
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

//// Main ////

func main() {
	conn, err := abrirDB()
	if err != nil {
		log.Fatalf("error al iniciar la BD: (%v)", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	server := handlers.NewServer(queries)

	http.HandleFunc("/", serveIndex)
	http.Handle("/users", middleware.Logging(http.HandlerFunc(server.UsersHandler)))
	http.Handle("/users/", middleware.Logging(http.HandlerFunc(server.UserHandler)))
	http.Handle("/library", middleware.Logging(http.HandlerFunc(server.LibrariesHandler)))
	http.Handle("/library/", middleware.Logging(http.HandlerFunc(server.LibraryHandler)))
	http.Handle("/catalogues/", middleware.Logging(http.HandlerFunc(server.CataloguesHandler)))
	http.Handle("/catalogue/", middleware.Logging(http.HandlerFunc(server.CatalogueHandler)))

	port := ":8080"
	fmt.Printf("Servidor escuchando en http://localhost%s\n", port)

	log.Fatal(http.ListenAndServe(port, nil))

}
