package models

type Receipt struct {
	ID           string `json:"id"`     // auto-generated, so no validation here
	Status       string `json:"status"` // PENDING, COMPLETED, or FAILED
	ErrorMessage string `json:"errorMessage,omitempty"`

	Retailer     string `json:"retailer" validate:"required"`
	PurchaseDate string `json:"purchaseDate" validate:"required"`
	PurchaseTime string `json:"purchaseTime" validate:"required"`
	Total        string `json:"total" validate:"required"`
	Items        []Item `json:"items" validate:"required,dive"` // "dive" ensures each item is validated
	Points       int    `json:"points"`                         // calculated later, so no validation
}
