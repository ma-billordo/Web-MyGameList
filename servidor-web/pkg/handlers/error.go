package handlers

import (
	"log"
	"net/http"

	"MyGameList.com/views"
)

// writeError responde con un status de error y el mensaje como fragmento HTML,
// para que HTMX lo muestre en el contenedor de error que corresponda:
// #form-error en los formularios, o el de cada tarjeta al borrar.
func WriteError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := views.ErrorMessage(msg).Render(r.Context(), w); err != nil {
		log.Printf("error al renderizar el mensaje de error: %v", err)
	}
}
