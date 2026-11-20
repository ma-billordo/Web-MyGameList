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
	//si está vacia devuelve nil y el for de templ simplemento no itera

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.IndexPage(gameList).Render(r.Context(), w); err != nil {
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
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Carga de datos en la BD
	_, err := s.queries.CreateGame(r.Context(), db.CreateGameParams{
		Title:       title,
		Description: description,
	})
	if err != nil {
		log.Printf("error al crear el juego: %v", err)
		http.Error(w, "error al crear el juego", http.StatusInternalServerError)
		return
	}

	// Redirect
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// POST /library/{id}/delete - Eliminar juego por id
func (s *Server) DeleteGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Capturar el ID y verificar DELETE del path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "invalid URL (length)", http.StatusBadRequest)
		return
	}
	if parts[3] != "delete" {
		http.Error(w, "invalid URL (path)", http.StatusBadRequest)
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
		http.Error(w, "error al eliminar el juego", http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "juego no encontrado", http.StatusNotFound)
		return
	}

	// Redirect
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
