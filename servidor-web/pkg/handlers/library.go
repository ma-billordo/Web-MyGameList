package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	db "MyGameList.com/db/sqlc"
)

func (s *Server) LibrariesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.ListGames(w, r)
	case http.MethodPost:
		s.CreateGame(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) LibraryHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}
	id64, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		http.Error(w, "invalid game ID", http.StatusBadRequest)
		return
	}
	id := int32(id64)
	// Elige según método
	switch r.Method {
	case http.MethodGet:
		s.GetGame(w, r, id)
	case http.MethodPut:
		s.UpdateGame(w, r, id)
	case http.MethodDelete:
		s.DeleteGame(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// POST /library - Crea nuevo juego
func (s *Server) CreateGame(w http.ResponseWriter, r *http.Request) {
	var newGame db.Library

	// Decodificación de datos en la variable
	err := json.NewDecoder(r.Body).Decode(&newGame)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validacion
	if msg := validateGameData(newGame.Title, newGame.Description); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Carga de datos en la BD
	gameID, err := s.queries.CreateGame(r.Context(), db.CreateGameParams{
		Title:       newGame.Title,
		Description: newGame.Description,
	})
	if err != nil {
		log.Printf("error al crear el juego: %v", err)
		http.Error(w, "error al crear el juego", http.StatusInternalServerError)
		return
	}

	// Codificación de datos
	newGame.IDGame = gameID
	newGame.AvgScore = "0.00"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newGame)
}

// GET /library - Listar juegos
func (s *Server) ListGames(w http.ResponseWriter, r *http.Request) {
	gameList, err := s.queries.ListGames(r.Context())
	if err != nil {
		log.Printf("error al listar los juegos: %v", err)
		http.Error(w, "error al listar los juegos", http.StatusInternalServerError)
		return
	}
	//si está vacia no devuelve null, devuelve la lista vacía
	if gameList == nil {
		gameList = []db.Library{} //crea un slice vacio de tipo Library
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(gameList)
}

// GET /library/{id} - Obtener juego por id
func (s *Server) GetGame(w http.ResponseWriter, r *http.Request, id int32) {
	game, err := s.queries.GetGame(r.Context(), id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "juego no encontrado", http.StatusNotFound)
			return
		}
		log.Printf("error al obtener el juego: %v", err)
		http.Error(w, "error al obtener el juego", http.StatusInternalServerError)
		return
	}

	//codificacion de datos
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(game)
}

// PUT /library/{id} - Actualizar juego por id
func (s *Server) UpdateGame(w http.ResponseWriter, r *http.Request, id int32) {
	var updatedGame db.UpdateGameParams

	// Decodificación de datos
	err := json.NewDecoder(r.Body).Decode(&updatedGame)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validación
	if msg := validateGameData(updatedGame.Title, updatedGame.Description); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Actualización de datos en la BD
	result, err := s.queries.UpdateGame(r.Context(), db.UpdateGameParams{
		IDGame:      id,
		Title:       updatedGame.Title,
		Description: updatedGame.Description,
	})

	if err != nil {
		log.Printf("error al actualizar el juego: %v", err)
		http.Error(w, "error al actualizar el juego", http.StatusInternalServerError)
		return
	}
	if result == 0 {
		http.Error(w, "juego no encontrado", http.StatusNotFound)
		return
	}

	// Codificación de datos
	updatedGame.IDGame = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(updatedGame)
}

// DELETE /library/{id} - Eliminar juego por id
func (s *Server) DeleteGame(w http.ResponseWriter, r *http.Request, id int32) {
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

	w.WriteHeader(http.StatusNoContent)
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
