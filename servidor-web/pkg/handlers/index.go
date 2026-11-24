package handlers

import (
	"log"
	"net/http"

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

	// Usuarios para el selector del catálogo
	userList, err := s.listUsersSorted(r.Context(), "asc")
	if err != nil {
		log.Printf("error al listar los usuarios: %v", err)
		http.Error(w, "error al listar los usuarios", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := views.IndexPage(userList).Render(r.Context(), w); err != nil {
		log.Printf("error al renderizar la pagina principal: %v", err)
	}
}
