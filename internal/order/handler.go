package order

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

type checkoutRequest struct {
	CardNumber string `json:"card_number"`
	ExpMonth   int    `json:"exp_month"`
	ExpYear    int    `json:"exp_year"`
	CVC        string `json:"cvc"`
}

type statusRequest struct {
	Status Status `json:"status"`
	Reason string `json:"reason"`
}

type createRequest struct {
	UserID int64       `json:"user_id"`
	Items  []ItemInput `json:"items"`
}

func (handler *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	var request checkoutRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}
	authenticatedUser, _ := auth.UserFromContext(r.Context())

	order, err := handler.service.Checkout(r.Context(), authenticatedUser.ID, Card{
		Number:   request.CardNumber,
		ExpMonth: request.ExpMonth,
		ExpYear:  request.ExpYear,
		CVC:      request.CVC,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, order)
}

func (handler *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	authenticatedUser, _ := auth.UserFromContext(r.Context())

	orders, err := handler.service.ListMine(r.Context(), authenticatedUser.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, orders)
}

func (handler *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	orders, err := handler.service.ListAll(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, orders)
}

func (handler *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var request statusRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	order, err := handler.service.UpdateStatus(r.Context(), r.PathValue("ref"), request.Status, request.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, order)
}

func (handler *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request createRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	order, err := handler.service.CreateForUser(r.Context(), request.UserID, request.Items)
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, order)
}

func writeError(w http.ResponseWriter, err error) {
	var validationErr ValidationError

	switch {
	case errors.As(err, &validationErr):
		httpjson.Error(w, http.StatusBadRequest, validationErr.Error())
	case errors.Is(err, ErrEmptyCart):
		httpjson.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrProductNotFound), errors.Is(err, ErrUserNotFound):
		httpjson.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrStatusChanged):
		httpjson.Error(w, http.StatusConflict, err.Error())
	default:
		log.Printf("order: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
	}
}
