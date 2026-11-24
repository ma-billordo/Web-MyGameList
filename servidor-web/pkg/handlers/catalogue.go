package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	db "MyGameList.com/db/sqlc"
	"MyGameList.com/views"
	"github.com/lib/pq"
)

// CatalogueHandler atiende /catalogue/...:
//   - /catalogue/{idUser}           -> POST crea un registro
//   - /catalogue/{idUser}/{idGame}  -> DELETE elimina un registro
func (s *Server) CatalogueHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")

	switch len(parts) {
	case 3:
		idUser, err := strconv.ParseInt(parts[2], 10, 32)
		if err != nil {
			http.Error(w, "invalid user ID", http.StatusBadRequest)
			return
		}
		s.CreateRecord(w, r, int32(idUser))
	case 4:
		idUser, err := strconv.ParseInt(parts[2], 10, 32)
		if err != nil {
			http.Error(w, "invalid user ID", http.StatusBadRequest)
			return
		}
		idGame, err := strconv.ParseInt(parts[3], 10, 32)
		if err != nil {
			http.Error(w, "invalid game ID", http.StatusBadRequest)
			return
		}
		s.DeleteRecord(w, r, int32(idUser), int32(idGame))
	default:
		http.Error(w, "invalid URL - URL expected: /catalogue/{idUser} or /catalogue/{idUser}/{idGame}", http.StatusBadRequest)
	}
}

// GET /catalogue?user={id} - Muestra el catálogo de un usuario
func (s *Server) CataloguePage(w http.ResponseWriter, r *http.Request) {
	// Capturar el ID del usuario; si falta o es inválido, se vuelve a la portada
	id64, err := strconv.ParseInt(r.URL.Query().Get("user"), 10, 32)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	idUser := int32(id64)

	// Verificar que el usuario exista (también se usa su nombre en el título)
	user, err := s.queries.GetUser(r.Context(), idUser)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "el usuario no existe", http.StatusNotFound)
			return
		}
		log.Printf("error al obtener el usuario: %v", err)
		http.Error(w, "error al obtener el usuario", http.StatusInternalServerError)
		return
	}

	// Juegos para el select del formulario
	games, err := s.queries.ListGames(r.Context())
	if err != nil {
		log.Printf("error al listar los juegos: %v", err)
		http.Error(w, "error al listar los juegos", http.StatusInternalServerError)
		return
	}

	// Registros del catálogo del usuario
	records, err := s.queries.ListCatalogueInfoByUser(r.Context(), idUser)
	if err != nil {
		log.Printf("error al listar el catálogo: %v", err)
		http.Error(w, "error al listar el catálogo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := views.CataloguePage(idUser, user.Name, games, records).Render(r.Context(), w); err != nil {
		log.Printf("error al renderizar la página: %v", err)
		http.Error(w, "error al mostrar la página", http.StatusInternalServerError)
	}
}

// POST /catalogues/{idUser} - Agregar juego al catálogo de un usuario
func (s *Server) CreateRecord(w http.ResponseWriter, r *http.Request, idUser int32) {
	// Leer campos
	status := r.FormValue("status")

	game64, err := strconv.ParseInt(r.FormValue("game"), 10, 32)
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, "juego inválido")
		return
	}
	idGame := int32(game64)

	// El puntaje es opcional: vacío = sin puntaje (nil)
	var score *int32
	if rawScore := r.FormValue("score"); rawScore != "" {
		score64, err := strconv.ParseInt(rawScore, 10, 32)
		if err != nil {
			WriteError(w, r, http.StatusBadRequest, "el puntaje debe ser un número")
			return
		}
		value := int32(score64)
		score = &value
	}

	// Validacion
	if msg := validateRecordData(status, score); msg != "" {
		WriteError(w, r, http.StatusBadRequest, msg)
		return
	}

	// Carga de datos en la BD
	err = s.queries.CreateRecord(r.Context(), db.CreateRecordParams{
		IDUser: idUser,
		IDGame: idGame,
		Score:  score,
		Status: status,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // unique_violation: clave primaria compuesta repetida
				WriteError(w, r, http.StatusConflict, "el juego ya está en el catálogo del usuario")
				return
			case "23503": // foreign_key_violation
				switch pqErr.Constraint {
				case "catalogue_users":
					WriteError(w, r, http.StatusNotFound, "el usuario no existe")
					return
				case "catalogue_library":
					WriteError(w, r, http.StatusBadRequest, "el juego no existe")
					return
				}
			}
		}
		log.Printf("error al crear el registro: %v", err)
		WriteError(w, r, http.StatusInternalServerError, "error al crear el registro")
		return
	}

	// Mostrar el catálogo actualizado (solo la lista)
	records, err := s.queries.ListCatalogueInfoByUser(r.Context(), idUser)
	if err != nil {
		log.Printf("error al listar el catálogo: %v", err)
		WriteError(w, r, http.StatusInternalServerError, "error al listar el catálogo")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := views.CatalogueList(records).Render(r.Context(), w); err != nil {
		log.Printf("error al renderizar el catálogo: %v", err)
	}
}

// DELETE /catalogue/{idUser}/{idGame} - Quita un juego del catálogo de un usuario
func (s *Server) DeleteRecord(w http.ResponseWriter, r *http.Request, idUser int32, idGame int32) {
	// Borrar de la BD
	rows, err := s.queries.DeleteRecord(r.Context(), db.DeleteRecordParams{
		IDUser: idUser,
		IDGame: idGame,
	})
	if err != nil {
		log.Printf("error al eliminar el registro: %v", err)
		WriteError(w, r, http.StatusInternalServerError, "error al eliminar el registro")
		return
	}
	if rows == 0 {
		WriteError(w, r, http.StatusNotFound, "registro no encontrado")
		return
	}

	// Respuesta vacía con 200: HTMX reemplaza el <article> por "nada"
	// (con 204 HTMX no intercambia nada y la tarjeta no se borraría)
	w.WriteHeader(http.StatusOK)
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
