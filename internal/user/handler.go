package user

import (
	"errors"
	"log"
	"net/http"
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
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     Role   `json:"role"`
}

type updateRequest struct {
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

func (handler *Handler) List(w http.ResponseWriter, r *http.Request) {
	users, err := handler.service.ListUsers(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, users)
}

func (handler *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request createRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	u, err := handler.service.CreateUser(r.Context(), request.Email, request.Password, request.Role)
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, u)
}

func (handler *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "id invalide")
		return
	}

	var request updateRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON invalide")
		return
	}

	u, err := handler.service.UpdateUser(r.Context(), id, request.Email, request.Role)
	if err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, u)
}

func (handler *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "id invalide")
		return
	}

	if err := handler.service.DeleteUser(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "id invalide")
		return
	}

	if err := handler.service.Confirm(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]string{"message": "compte confirmé"})
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func writeError(w http.ResponseWriter, err error) {
	var validationErr ValidationError
	var notFoundErr NotFoundError
	var emailTakenErr EmailTakenError

	switch {
	case errors.As(err, &validationErr):
		httpjson.Error(w, http.StatusBadRequest, validationErr.Error())
	case errors.As(err, &notFoundErr):
		httpjson.Error(w, http.StatusNotFound, notFoundErr.Error())
	case errors.As(err, &emailTakenErr):
		httpjson.Error(w, http.StatusConflict, emailTakenErr.Error())
	default:
		log.Printf("user: %v", err)
		httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
	}
}
