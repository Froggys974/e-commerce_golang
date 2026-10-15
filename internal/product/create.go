package product

import (
	"log"
	"net/http"
	"strings"

	"github.com/Froggys974/e-commerce_golang/internal/db"
	"github.com/Froggys974/e-commerce_golang/internal/httpjson"
	"github.com/Froggys974/e-commerce_golang/internal/ref"
)

const (
	defaultVATPercent = 20
	maxRefTries       = 5
)

type createRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	PriceHTCents int64  `json:"price_ht_cents"`
	VATPercent   *int   `json:"vat_percent"` // pointeur : nil = champ pas dans le json
}

func (handler *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request createRequest
	if err := httpjson.Read(w, r, &request); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "JSON non valide")
		return
	}

	name := strings.TrimSpace(request.Name)
	description := strings.TrimSpace(request.Description)
	category := strings.ToLower(strings.TrimSpace(request.Category))

	if name == "" || category == "" {
		httpjson.Error(w, http.StatusBadRequest, "name et category manquants, obligatoires")
		return
	}
	if request.PriceHTCents < 0 {
		httpjson.Error(w, http.StatusBadRequest, "prix non valide")
		return
	}

	vatPercent := defaultVATPercent
	if request.VATPercent != nil {
		vatPercent = *request.VATPercent
	}
	if vatPercent < 0 || vatPercent > 100 {
		httpjson.Error(w, http.StatusBadRequest, "TVA invalide (0 à 100)")
		return
	}

	// si même ref quand même on reessaie (peu probable mais bon)
	for attempt := 1; attempt <= maxRefTries; attempt++ {
		productRef := ref.New(ref.Product)

		_, err := handler.database.Exec(
			`INSERT INTO products (ref, name, description, category, price_ht_cents, vat_percent)
                           VALUES ($1, $2, $3, $4, $5, $6)`,
			productRef, name, description, category, request.PriceHTCents, vatPercent,
		)
		if db.IsUniqueViolation(err) {
			log.Printf("product create: collision ref %s, tentative %d", productRef, attempt)
			continue
		}
		if err != nil {
			log.Printf("product create: insert: %v", err)
			httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
			return
		}

		httpjson.Write(w, http.StatusCreated, Product{
			Ref:           productRef,
			Name:          name,
			Description:   description,
			Category:      category,
			PriceHTCents:  request.PriceHTCents,
			VATPercent:    vatPercent,
			PriceTTCCents: PriceTTC(request.PriceHTCents, vatPercent),
		})
		return
	}

	log.Printf("product create: %d collisions de ref d'affilée", maxRefTries)
	httpjson.Error(w, http.StatusInternalServerError, "Erreur serveur")
}
