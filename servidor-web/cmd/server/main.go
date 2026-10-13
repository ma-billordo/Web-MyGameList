package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	db "MyGameList.com/db/sqlc"

	_ "github.com/lib/pq"
)

//// Funciones /////

func abrirDB() (*sql.DB, error) {
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		return nil, fmt.Errorf("La variable de entorno DB_URL no está seteada")
	}
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		defer conn.Close()
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		defer conn.Close()
		return nil, err
	}
	conn.SetMaxOpenConns(25)
	return conn, nil
}

//// Handlers ////

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

//// Main ////

func main() {
	conn, err := abrirDB()
	if err != nil {
		log.Fatalf("Error al iniciar la BD: (%v)", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	ctx := context.Background()

	//------ main anterior
	http.HandleFunc("/", serveForm)

	port := ":8080"
	fmt.Printf("Servidor escuchando en http://localhost%s\n", port)

	err = http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Printf("Error al iniciar el servidor: %s\n", err)
	}
}
