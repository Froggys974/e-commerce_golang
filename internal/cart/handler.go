package cart

import (
	"errors"
	"log"
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/auth"
	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type setItemRequest struct {
	Quantity int `json:"quantity"`
}

func (handler *Handler) Get(w http.ResponseWriter, r *http.Request) {
	authenticatedUser, _ := auth.UserFromContext(r.Context())

	cart, err := handler.service.Get(r.Context(), authenticatedUser.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, cart)
}

func (handler *Handler) SetItem(w http.ResponseWriter, r *http.Request) {
	var request setItemRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}
	authenticatedUser, _ := auth.UserFromContext(r.Context())

	cart, err := handler.service.SetItem(r.Context(), authenticatedUser.ID, r.PathValue("ref"), request.Quantity)
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, cart)
}

func (handler *Handler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	authenticatedUser, _ := auth.UserFromContext(r.Context())

	cart, err := handler.service.RemoveItem(r.Context(), authenticatedUser.ID, r.PathValue("ref"))
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, cart)
}

func writeError(w http.ResponseWriter, err error) {
	var validationErr ValidationError

	switch {
	case errors.As(err, &validationErr):
		httpjson.Error(w, http.StatusBadRequest, validationErr.Error())
	case errors.Is(err, ErrProductNotFound), errors.Is(err, ErrItemNotFound):
		httpjson.Error(w, http.StatusNotFound, err.Error())
	default:
		log.Printf("cart: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
	}
}
