package client

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
	"github.com/Froggys974/e-commerce_golang/internal/order"
)

type checkoutDoneMsg struct {
	order order.Order
	err   error
}

// key pour relire les valeurs car le modele est copie a chaque update
func newCheckoutForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("card_number").
				Title("Numéro de carte").
				Placeholder("4242 4242 4242 4242").
				Validate(required),
			huh.NewInput().
				Key("exp_month").
				Title("Mois d'expiration").
				Placeholder("12").
				Validate(isNumber),
			huh.NewInput().
				Key("exp_year").
				Title("Année d'expiration").
				Placeholder("2030").
				Validate(isNumber),
			huh.NewInput().
				Key("cvc").
				Title("CVC").
				EchoMode(huh.EchoModePassword).
				Validate(required),
		),
	)
}

func checkoutCmd(api *apiclient.Client, card apiclient.CheckoutRequest) tea.Cmd {
	return func() tea.Msg {
		created, err := api.Checkout(card)
		return checkoutDoneMsg{order: created, err: err}
	}
}

func (model catalog) updateCheckout(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok && key.String() == "esc" {
		model.checkout = nil
		return model, nil
	}

	updated, cmd := model.checkout.Update(msg)
	if form, ok := updated.(*huh.Form); ok {
		model.checkout = form
	}

	switch model.checkout.State {
	case huh.StateAborted:
		model.checkout = nil
		return model, nil
	case huh.StateCompleted:
		// deja verifie par isNumber
		month, _ := strconv.Atoi(strings.TrimSpace(model.checkout.GetString("exp_month")))
		year, _ := strconv.Atoi(strings.TrimSpace(model.checkout.GetString("exp_year")))
		card := apiclient.CheckoutRequest{
			CardNumber: model.checkout.GetString("card_number"),
			ExpMonth:   month,
			ExpYear:    year,
			CVC:        model.checkout.GetString("cvc"),
		}
		model.checkout = nil
		model.status = "Paiement en cours..."
		return model, checkoutCmd(model.api, card)
	}
	return model, cmd
}

func (model catalog) applyCheckout(msg checkoutDoneMsg) catalog {
	model.err = msg.err
	model.status = ""
	if msg.err != nil {
		return model
	}

	model.status = fmt.Sprintf("✓ Commande %s payée • total %s", msg.order.Ref, formatPrice(msg.order.TotalTTCCents))
	model.cart.Items = nil
	model.cart.TotalTTCCents = 0
	model.cartTable.SetRows(nil)
	return model
}

func (model catalog) checkoutView() string {
	var builder strings.Builder
	builder.WriteString(titleStyle.Render("Paiement"))
	builder.WriteString("\n\n")
	fmt.Fprintf(&builder, "Montant : %s\n\n", formatPrice(model.cart.TotalTTCCents))
	builder.WriteString(model.checkout.View())
	builder.WriteString("\n\n")
	builder.WriteString(helpStyle.Render("échap annuler"))
	return builder.String()
}

func isNumber(value string) error {
	if _, err := strconv.Atoi(strings.TrimSpace(value)); err != nil {
		return errors.New("doit être un nombre")
	}
	return nil
}
