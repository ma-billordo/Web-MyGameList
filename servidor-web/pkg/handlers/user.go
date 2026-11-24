package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	db "MyGameList.com/db/sqlc"
	"MyGameList.com/views"
	"github.com/lib/pq"
)

// UsersHandler atiende /users: GET muestra la página y POST crea un usuario.
func (s *Server) UsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.UsersPage(w, r)
	case http.MethodPost:
		s.CreateUser(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Carga de la página de usuarios
func (s *Server) UsersPage(w http.ResponseWriter, r *http.Request) {
	userOrder := sortOrderFromURL(r.URL.String())
	userList, err := s.listUsersSorted(r.Context(), userOrder)
	if err != nil {
		log.Printf("error al listar los usuarios: %v", err)
		http.Error(w, "error al listar los usuarios", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Pedido de HTMX (clic en la cabecera): solo la tabla
	if isPartialRequest(r) {
		if err := views.UserList(userList, userOrder).Render(r.Context(), w); err != nil {
			log.Printf("error al renderizar la tabla de usuarios: %v", err)
		}
		return
	}

	if err := views.UserPage(userList, userOrder).Render(r.Context(), w); err != nil {
		log.Printf("error al renderizar la página: %v", err)
		http.Error(w, "error al mostrar la página", http.StatusInternalServerError)
	}
}

// POST /users - Crea nuevo usuario
func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Leer campos
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Validacion
	if msg := validateUserData(name, email, password); msg != "" {
		WriteError(w, r, http.StatusBadRequest, msg)
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
			WriteError(w, r, http.StatusConflict, "ya existe un usuario con ese email")
			return
		}
		log.Printf("error al crear el usuario: %v", err)
		WriteError(w, r, http.StatusInternalServerError, "error al crear el usuario")
		return
	}

	// Mostrar los usuarios
	// Respetar el orden que el usuario está viendo (viene en la URL actual del navegador)
	order := sortOrderFromURL(r.Header.Get("HX-Current-URL"))
	userList, err := s.listUsersSorted(r.Context(), order)
	if err != nil {
		log.Printf("error al listar los usuarios: %v", err)
		WriteError(w, r, http.StatusInternalServerError, "error al listar los usuarios")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := views.UserList(userList, order).Render(r.Context(), w); err != nil {
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

	// Borrar de la BD
	rows, err := s.queries.DeleteUser(r.Context(), id)
	if err != nil {
		log.Printf("error al eliminar el usuario: %v", err)
		http.Error(w, "no se pudo eliminar el usuario", http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "el usuario no existe (puede que ya se haya borrado)", http.StatusNotFound)
		return
	}
	// Respuesta vacía con 200
	w.WriteHeader(http.StatusOK)
}

// Funciones auxiliares

// isPartialRequest indica si el pedido es solo para actualizar una parte de la página
// (por ejemplo, el clic en la cabecera de la tabla) y no una navegación completa.
func isPartialRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true"
}

// validateUserData valida "name", "email" y "password" del Create de un usuario.
// Devuelve el mensaje de error, o "" si los datos son válidos.
func validateUserData(name string, email string, password string) string {
	// Validación name
	if strings.TrimSpace(name) == "" {
		return "el nombre no puede estar vacío"
	}
	if len([]rune(name)) > 255 {
		return "el nombre no puede superar los 255 caracteres"
	}
	// Validación email
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if !re.MatchString(email) {
		return "formato de email inválido"
	}
	if len([]rune(email)) > 255 {
		return "el email no puede superar los 255 caracteres"
	}
	// Validación password
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

// sortOrderFromURL devuelve "desc" si la URL tiene order=desc, o "asc" en cualquier otro caso.
func sortOrderFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Query().Get("order") != "desc" {
		return "asc"
	}
	return "desc"
}

// listUsersSorted trae los usuarios ordenados por nombre, ascendente o descendente.
func (s *Server) listUsersSorted(ctx context.Context, order string) ([]db.UserInfo, error) {
	switch order {
	case "desc":
		return s.queries.ListUsersOrderByNameDesc(ctx)
	default:
		return s.queries.ListUsersOrderByNameAsc(ctx)
	}
}
