package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	db "MyGameList.com/db/sqlc"
	"MyGameList.com/views"
)

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	gameList, err := s.queries.ListGames(r.Context())
	if err != nil {
		log.Printf("error al listar los juegos: %v", err)
		http.Error(w, "error al listar los juegos", http.StatusInternalServerError)
		return
	}

	//agregado nuevo (usuarios)
	userList, err := s.queries.ListUsers(r.Context())
	if err != nil {
		log.Printf("error al listar los usuarios: %v", err)
		http.Error(w, "error al listar los usuarios", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.IndexPage(gameList, userList).Render(r.Context(), w); err != nil {
		log.Printf("error al renderizar la página: %v", err)
		http.Error(w, "error al mostrar la página", http.StatusInternalServerError)
	}
}

// POST /library - Crea nuevo juego
func (s *Server) CreateGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Leer campos
	title := r.FormValue("title")
	description := r.FormValue("description")

	// Validacion
	if msg := validateGameData(title, description); msg != "" {
		writeError(w, r, http.StatusBadRequest, msg)
		return
	}

	// Carga de datos en la BD
	_, err := s.queries.CreateGame(r.Context(), db.CreateGameParams{
		Title:       title,
		Description: description,
	})
	if err != nil {
		log.Printf("error al crear el juego: %v", err)
		writeError(w, r, http.StatusInternalServerError, "error al crear el juego")
		return
	}

	gameList, err := s.queries.ListGames(r.Context())
	if err != nil {
		log.Printf("error al listar los juegos: %v", err)
		writeError(w, r, http.StatusInternalServerError, "error al listar los juegos")
		return
	}
	//si está vacia devuelve nil y el for de templ simplemento no itera

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := views.LibraryList(gameList).Render(r.Context(), w); err != nil {
		log.Printf("error al renderizar la lista: %v", err)
		http.Error(w, "error al mostrar la lista", http.StatusInternalServerError)
		return
	}
}

// DELETE /library/{id} - Eliminar juego por id
func (s *Server) DeleteGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Capturar el ID
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		http.Error(w, "invalid URL (length)", http.StatusBadRequest)
		return
	}
	id64, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		http.Error(w, "invalid game ID", http.StatusBadRequest)
		return
	}
	id := int32(id64)

	// Borrar de la BD
	rows, err := s.queries.DeleteGame(r.Context(), id)
	if err != nil {
		log.Printf("error al eliminar el juego: %v", err)
		writeError(w, r, http.StatusInternalServerError, "error al eliminar el juego")
		return
	}
	if rows == 0 {
		writeError(w, r, http.StatusNotFound, "juego no encontrado")
		return
	}

	// Respuesta vacía con 200: HTMX reemplaza el <article> por "nada"
	w.WriteHeader(http.StatusOK)
}

// Funciones auxiliares

// validateGameData valida "title" y "description" del Create y Update de un juego.
// Devuelve el mensaje de error, o "" si los datos son válidos.
func validateGameData(title string, description string) string {
	if strings.TrimSpace(title) == "" { //trimSpace quita los espacios de los extremos. "   " cuenta como vacío.
		return "el título del juego no puede estar vacío"
	}
	if len([]rune(title)) > 255 { // len cuenta bytes y []rune hace que cuente caracteres (ñ o ´ ocupa 2B.).
		return "el título del juego no puede superar los 255 caracteres"
	}
	if strings.TrimSpace(description) == "" {
		return "la descripción del juego no puede estar vacía"
	}
	return ""
}
