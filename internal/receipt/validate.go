package receipt

import (
	"regexp"
	"time"

	"github.com/kartikeya55555/fetch-assignment/internal/errors"
	"github.com/kartikeya55555/fetch-assignment/internal/models"
)

func validateReceipt(r *models.Receipt) []string {
	var issues []string

	// Validate PurchaseDate
	if _, err := time.Parse("2006-01-02", r.PurchaseDate); err != nil {
		issues = append(issues, errors.ErrNotValidDateFormat.Error())
	}

	// Validate PurchaseTime
	if _, err := time.Parse("15:04", r.PurchaseTime); err != nil {
		issues = append(issues, errors.ErrNotValidTimeFormat.Error())
	}

	// Validate total format (\d+\.\d{2})
	matched, _ := regexp.MatchString(`^\d+\.\d{2}$`, r.Total)
	if !matched {
		issues = append(issues, errors.ErrNotValidTotalFormat.Error())
	}

	return issues
}
