package receipt

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/kartikeya55555/fetch-assignment/internal/models"
)

type PointsCalculator interface {
	CalculatePoints(r *models.Receipt) int
}

type defaultPointsCalculator struct{}

func NewDefaultPointsCalculator() PointsCalculator {
	return &defaultPointsCalculator{}
}

// CalculatePoints implements the same logic you had before
func (calc *defaultPointsCalculator) CalculatePoints(r *models.Receipt) int {
	points := 0

	// 1) One point per alphanumeric in retailer
	for _, ch := range r.Retailer {
		if isAlphanumeric(ch) {
			points++
		}
	}

	totalF, _ := strconv.ParseFloat(r.Total, 64)

	// 2) +50 if total is round dollar
	if hasNoCents(totalF) {
		points += 50
	}

	// 3) +25 if total is multiple of 0.25
	if isMultipleOfQuarter(totalF) {
		points += 25
	}

	// 4) +5 points for every 2 items
	points += (len(r.Items) / 2) * 5

	// 5) If desc len %3 == 0 => +ceil(price * 0.2)
	for _, item := range r.Items {
		desc := strings.TrimSpace(item.ShortDescription)
		if len(desc)%3 == 0 {
			pF, _ := strconv.ParseFloat(item.Price, 64)
			points += int(math.Ceil(pF * 0.2))
		}
	}

	// 6) +6 if purchase day is odd
	d, _ := time.Parse("2006-01-02", r.PurchaseDate)
	if d.Day()%2 == 1 {
		points += 6
	}

	// 7) +10 if purchase time is after 2pm and before 4pm
	t, _ := time.Parse("15:04", r.PurchaseTime)
	if t.Hour() == 14 && t.Minute() > 0 || (t.Hour() > 14 && t.Hour() < 16) {
		points += 10
	}

	return points
}

// Helpers
func isAlphanumeric(ch rune) bool {
	return (ch >= '0' && ch <= '9') ||
		(ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z')
}

func hasNoCents(f float64) bool {
	return f == float64(int64(f))
}

func isMultipleOfQuarter(f float64) bool {
	rem := math.Mod(f, 0.25)
	return math.Abs(rem) < 1e-9
}
