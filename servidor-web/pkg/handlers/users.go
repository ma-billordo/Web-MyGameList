package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	db "MyGameList.com/db/sqlc"
	"MyGameList.com/views"
	"github.com/lib/pq"
)

// POST /users - Crea nuevo usuario
func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Leer campos
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Validacion
	if msg := validateUserData(name, email, password); msg != "" {
		writeError(w, r, http.StatusBadRequest, msg)
		return
	}

	// Carga de datos en la BD
	_, err := s.queries.CreateUser(r.Context(), db.CreateUserParams{
		Name:     name,
		Email:    email,
		Password: password,
	})
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, r, http.StatusBadRequest, "ya existe un usuario con ese email")
			return
		}
		log.Printf("error al crear el usuario: %v", err)
		writeError(w, r, http.StatusInternalServerError, "error al crear el usuario")
		return
	}

	userList, err := s.queries.ListUsers(r.Context())
	if err != nil {
		log.Printf("error al listar los usuarios: %v", err)
		writeError(w, r, http.StatusInternalServerError, "error al listar los usuarios")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := views.UserList(userList).Render(r.Context(), w); err != nil {
		log.Printf("error al renderizar la lista de usuarios: %v", err)
	}
}

// DELETE /users/{id} - Eliminar usuario por id
func (s *Server) DeleteUser(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	id := int32(id64)

	// Borrar de la BD (por el ON DELETE CASCADE también se borra su Catalogue)
	rows, err := s.queries.DeleteUser(r.Context(), id)
	if err != nil {
		log.Printf("error al eliminar el usuario: %v", err)
		writeError(w, r, http.StatusInternalServerError, "error al eliminar el usuario")
		return
	}
	if rows == 0 {
		writeError(w, r, http.StatusNotFound, "usuario no encontrado")
		return
	}

	// Respuesta vacía con 200: HTMX reemplaza el <article> por "nada"
	w.WriteHeader(http.StatusOK)
}

// Funciones auxiliares

// validateUserData valida "name", "email" y "password" del Create de un usuario.
// Devuelve el mensaje de error, o "" si los datos son válidos.
func validateUserData(name string, email string, password string) string {
	if strings.TrimSpace(name) == "" {
		return "el nombre no puede estar vacío"
	}
	if len([]rune(name)) > 255 {
		return "el nombre no puede superar los 255 caracteres"
	}
	if strings.TrimSpace(email) == "" {
		return "el email no puede estar vacío"
	}
	if len([]rune(email)) > 255 {
		return "el email no puede superar los 255 caracteres"
	}
	if !strings.Contains(email, "@") {
		return "el email no es válido"
	}
	if strings.TrimSpace(password) == "" {
		return "la contraseña no puede estar vacía"
	}
	return ""
}

// isUniqueViolation indica si el error de Postgres es por violar una restricción UNIQUE
// (en Users, el email repetido). 23505 es el código de Postgres para "unique_violation".
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

// writeError responde con un status de error y el mensaje como fragmento HTML,
// para que HTMX lo muestre en el contenedor de error que corresponda
// (#form-error en el formulario, o el de cada tarjeta al borrar).
func writeError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := views.ErrorMessage(msg).Render(r.Context(), w); err != nil {
		log.Printf("error al renderizar el mensaje de error: %v", err)
	}
}
