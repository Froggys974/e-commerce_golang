package product

import (
	"errors"
	"log"
	"net/http"

	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type createRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	PriceHTCents int64  `json:"price_ht_cents"`
	VATPercent   *int   `json:"vat_percent"` // nil si absent du json
}

func (handler *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request createRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	p, err := handler.service.Create(r.Context(), CreateInput{
		Name:         request.Name,
		Description:  request.Description,
		Category:     request.Category,
		PriceHTCents: request.PriceHTCents,
		VATPercent:   request.VATPercent,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, p)
}

func writeError(w http.ResponseWriter, err error) {
	var validationErr ValidationError

	switch {
	case errors.As(err, &validationErr):
		httpjson.Error(w, http.StatusBadRequest, validationErr.Error())
	default:
		log.Printf("product: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
	}
}
