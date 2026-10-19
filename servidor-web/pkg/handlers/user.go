package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	db "MyGameList.com/db/sqlc"
	"github.com/lib/pq"
)

func (s *Server) UsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.ListUsers(w, r)
	case http.MethodPost:
		s.CreateUser(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
func (s *Server) UserHandler(w http.ResponseWriter, r *http.Request) {
	// Extraer el ID del path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}
	id64, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	id := int32(id64)
	// Elige según método
	switch r.Method {
	case http.MethodGet:
		s.GetUser(w, r, id)
	case http.MethodPut:
		s.UpdateUser(w, r, id)
	case http.MethodDelete:
		s.DeleteUser(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// POST /users - Crea nuevo usuario
func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var newUser db.User

	// Decodificación de datos en la variable
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validación
	if newUser.Name == "" {
		http.Error(w, "el nombre del usuario no puede estar vacío", http.StatusBadRequest)
		return
	}
	if err := validateEmail(newUser.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(newUser.Password) < 4 {
		http.Error(w, "la contraseña del usuario debe tener al menos 4 dígitos", http.StatusBadRequest)
		return
	}

	// Carga de datos en la BD
	userID, err := s.queries.CreateUser(r.Context(), db.CreateUserParams{
		Name:     newUser.Name,
		Email:    newUser.Email,
		Password: newUser.Password,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			http.Error(w, "email ya en uso", http.StatusConflict)
			return
		}
		log.Printf("error al crear usuario: %v", err)
		http.Error(w, "error al crear usuario", http.StatusInternalServerError)
		return
	}

	// Codificación de datos
	newUser.IDUser = userID
	newUser.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newUser)
}

// GET /users - Listar usuarios
func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
	userlist, err := s.queries.ListUsers(r.Context())
	if err != nil {
		log.Printf("error al listar los usuarios: %v", err)
		http.Error(w, "error al listar los usuarios", http.StatusInternalServerError)
		return
	}
	//si está vacia no devuelve null, devuelve la lista vacía
	if userlist == nil {
		userlist = []db.ListUsersRow{} //crea un slice vacio de tipo ListUsersRow
	}

	// Codificación de datos
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(userlist)
}

// GET /users/{id} - Obtener usuario por id
func (s *Server) GetUser(w http.ResponseWriter, r *http.Request, id int32) {
	user, err := s.queries.GetUser(r.Context(), id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "usuario no encontrado", http.StatusNotFound)
			return
		}
		log.Printf("error al obtener el usuario: %v", err)
		http.Error(w, "error al obtener el usuario", http.StatusInternalServerError)
		return
	}

	// Codificación de datos
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(user)
}

// PUT /users/{id} - Actualizar usuario por id
func (s *Server) UpdateUser(w http.ResponseWriter, r *http.Request, id int32) {
	var updatedUser db.UpdateUserParams
	err := json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validación
	if updatedUser.Name == "" {
		http.Error(w, "el nombre nuevo del usuario no puede estar vacío", http.StatusBadRequest)
		return
	}
	if err := validateEmail(updatedUser.Email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Actualización de datos en la BD
	result, err := s.queries.UpdateUser(r.Context(), db.UpdateUserParams{
		IDUser: id,
		Name:   updatedUser.Name,
		Email:  updatedUser.Email,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			http.Error(w, "email ya en uso", http.StatusConflict)
			return
		}
		log.Printf("error al actualizar el usuario: %v", err)
		http.Error(w, "error al actualizar el usuario", http.StatusInternalServerError)
		return
	}
	if result == 0 {
		http.Error(w, "usuario no encontrado", http.StatusNotFound)
		return
	}

	// Codificación de datos
	updatedUser.IDUser = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(updatedUser)
}

// DELETE /users/{id} - Eliminar usuario por id
func (s *Server) DeleteUser(w http.ResponseWriter, r *http.Request, id int32) {
	rows, err := s.queries.DeleteUser(r.Context(), id)
	if err != nil {
		log.Printf("error al eliminar el usuario: %v", err)
		http.Error(w, "error al eliminar el usuario", http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "usuario no encontrado", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Funciones auxiliares
func validateEmail(email string) error {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if !re.MatchString(email) {
		return errors.New("formato de email inválido")
	}

	return nil
}
