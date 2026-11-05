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
	"github.com/lib/pq"
)

func (s *Server) CataloguesHandler(w http.ResponseWriter, r *http.Request) {
	// Extraer el ID del usuario del path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		http.Error(w, "invalid URL - URL expected: /catalogues/{idUser}", http.StatusBadRequest)
		return
	}
	id64, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	idUser := int32(id64)

	switch r.Method {
	case http.MethodGet:
		s.ListRecord(w, r, idUser)
	case http.MethodPost:
		s.CreateRecord(w, r, idUser)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) CatalogueHandler(w http.ResponseWriter, r *http.Request) {
	// Extraer el ID de usuario e ID del juego del path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "invalid URL - URL expected: /catalogue/{idUser}/{idGame}", http.StatusBadRequest)
		return
	}
	idU64, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	idG64, err := strconv.ParseInt(parts[3], 10, 32)
	if err != nil {
		http.Error(w, "invalid game ID", http.StatusBadRequest)
		return
	}
	idUser := int32(idU64)
	idGame := int32(idG64)
	// Elige según método
	switch r.Method {
	case http.MethodGet:
		s.GetRecord(w, r, idUser, idGame)
	case http.MethodPut:
		s.UpdateRecord(w, r, idUser, idGame)
	case http.MethodDelete:
		s.DeleteRecord(w, r, idUser, idGame)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// POST /catalogues/{idUser} - Agregar juego al catálogo de un usuario
func (s *Server) CreateRecord(w http.ResponseWriter, r *http.Request, idUser int32) {
	var newRecord db.Catalogue

	// Decodificacion
	err := json.NewDecoder(r.Body).Decode(&newRecord)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validación
	if newRecord.IDGame <= 0 {
		http.Error(w, "id del juego ingresado invalido", http.StatusBadRequest)
		return
	}
	if msg := validateRecordData(newRecord.Status, newRecord.Score); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// cargar datos en la BD
	err = s.queries.CreateRecord(r.Context(), db.CreateRecordParams{
		IDUser: idUser,
		IDGame: newRecord.IDGame,
		Score:  newRecord.Score,
		Status: newRecord.Status,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // unique_violation: clave primaria compuesta repetida
				http.Error(w, "el juego ya está en el catálogo del usuario", http.StatusConflict)
				return
			case "23503": // foreign_key_violation
				switch pqErr.Constraint {
				case "catalogue_users":
					http.Error(w, "el usuario no existe", http.StatusNotFound)
					return
				case "catalogue_library":
					http.Error(w, "el juego no existe", http.StatusBadRequest)
					return
				}
			}
		}
		log.Printf("error al crear registro: %v", err)
		http.Error(w, "error al crear registro", http.StatusInternalServerError)
		return
	}

	// Codificación de datos
	newRecord.IDUser = idUser
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newRecord)
}

// GET /catalogues/{idUser} - Listar juegos del catálogo de un usuario
func (s *Server) ListRecord(w http.ResponseWriter, r *http.Request, idUser int32) {
	//verifica si el usuario existe primero
	_, err := s.queries.GetUser(r.Context(), idUser)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "el usuario no existe", http.StatusNotFound)
			return
		}
		log.Printf("error al obtener el usuario: %v", err)
		http.Error(w, "error al obtener el usuario", http.StatusInternalServerError)
		return
	}

	// Obtencion de datos
	recordList, err := s.queries.ListRecord(r.Context(), idUser)
	if err != nil {
		log.Printf("error al listar los registros: %v", err)
		http.Error(w, "error al listar los registros", http.StatusInternalServerError)
		return
	}
	//si está vacia no devuelve null, devuelve la lista vacía
	if recordList == nil {
		recordList = []db.Catalogue{} //crea un slice vacio de tipo Catalogue
	}

	//Codificacion de datos
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(recordList)
}

// GET /catalogue/{idUser}/{idGame} - Obtener un registro específico del catálogo de un usuario
func (s *Server) GetRecord(w http.ResponseWriter, r *http.Request, idUser int32, idGame int32) {
	record, err := s.queries.GetRecord(r.Context(), db.GetRecordParams{
		IDUser: idUser,
		IDGame: idGame,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "registro no encontrado", http.StatusNotFound)
			return
		}
		log.Printf("error al obtener el registro: %v", err)
		http.Error(w, "error al obtener el registro", http.StatusInternalServerError)
		return
	}

	//Codificacion
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(record)
}

// PUT /catalogue/{idUser}/{idGame} - Actualizar un registro del catálogo de un usuario
func (s *Server) UpdateRecord(w http.ResponseWriter, r *http.Request, idUser int32, idGame int32) {
	var updatedRecord db.UpdateRecordParams

	// Decodificion de datos
	err := json.NewDecoder(r.Body).Decode(&updatedRecord)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//validacion
	if msg := validateRecordData(updatedRecord.Status, updatedRecord.Score); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// actualizar en la BD
	result, err := s.queries.UpdateRecord(r.Context(), db.UpdateRecordParams{
		IDUser: idUser,
		IDGame: idGame,
		Score:  updatedRecord.Score,
		Status: updatedRecord.Status,
	})

	if err != nil {
		log.Printf("error al actualizar el registro: %v", err)
		http.Error(w, "error al actualizar el registro", http.StatusInternalServerError)
		return
	}
	if result == 0 {
		http.Error(w, "registro no encontrado", http.StatusNotFound)
		return
	}

	// Codificación de datos
	updatedRecord.IDUser = idUser
	updatedRecord.IDGame = idGame
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(updatedRecord)
}

// DELETE /catalogue/{idUser}/{idGame} - Eliminar un registro del catálogo de un usuario
func (s *Server) DeleteRecord(w http.ResponseWriter, r *http.Request, idUser int32, idGame int32) {
	rows, err := s.queries.DeleteRecord(r.Context(), db.DeleteRecordParams{
		IDUser: idUser,
		IDGame: idGame,
	})
	if err != nil {
		log.Printf("error al eliminar el registro: %v", err)
		http.Error(w, "error al eliminar el registro", http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "registro no encontrado", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/// Funciones auxiliares

// validateRecordData valida los campos que comparten el Create y el Update.
// Devuelve el mensaje de error, o "" si los datos son válidos.
func validateRecordData(status string, score *int32) string {
	switch status {
	case "Por jugar", "Jugando", "Completado", "Platinado":
		// estado válido
	default:
		return "estado inválido"
	}

	if score != nil && (*score < 1 || *score > 5) {
		return "el puntaje debe estar entre 1 y 5"
	}

	return ""
}
