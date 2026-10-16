package product

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"

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

	product, err := handler.service.Create(r.Context(), CreateInput{
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
	httpjson.Write(w, http.StatusCreated, product)
}

// ex GET /products?name=&description=&category=&min_price_ht=&max_price_ht=&min_price_ttc=&max_price_ttc=
func (handler *Handler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	minPriceHT, errMinPriceHT := parseCentsParam(query, "min_price_ht")
	maxPriceHT, errMaxPriceHT := parseCentsParam(query, "max_price_ht")
	minPriceTTC, errMinPriceTTC := parseCentsParam(query, "min_price_ttc")
	maxPriceTTC, errMaxPriceTTC := parseCentsParam(query, "max_price_ttc")
	// Join renvoie nil si toutes les erreurs sont nil
	if err := errors.Join(errMinPriceHT, errMaxPriceHT, errMinPriceTTC, errMaxPriceTTC); err != nil {
		writeError(w, err)
		return
	}

	products, err := handler.service.Search(r.Context(), SearchFilters{
		Name:        query.Get("name"),
		Description: query.Get("description"),
		Category:    query.Get("category"),
		MinPriceHT:  minPriceHT,
		MaxPriceHT:  maxPriceHT,
		MinPriceTTC: minPriceTTC,
		MaxPriceTTC: maxPriceTTC,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, products)
}

// GET /products/{ref}
func (handler *Handler) Get(w http.ResponseWriter, r *http.Request) {
	product, err := handler.service.GetByRef(r.Context(), r.PathValue("ref"))
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, product)
}

func parseCentsParam(query url.Values, param string) (*int64, error) {
	rawValue := query.Get(param)
	if rawValue == "" {
		return nil, nil
	}
	cents, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		return nil, ValidationError{Field: param, Message: "doit être un entier (centimes)"}
	}
	return &cents, nil
}

func writeError(w http.ResponseWriter, err error) {
	var validationErr ValidationError

	switch {
	case errors.As(err, &validationErr):
		httpjson.Error(w, http.StatusBadRequest, validationErr.Error())
	case errors.Is(err, ErrNotFound):
		httpjson.Error(w, http.StatusNotFound, err.Error())
	default:
		log.Printf("product: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
	}
}
