package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne-app/internal/models"
	"fyne-app/internal/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Helper struct for displaying purchase/sell items with full details
type DisplayItem struct {
	Code  string
	Name  string
	Qty   string
	Price string
	Total string
}

// FormatCurrency formats a float64 to Rupiah currency string
func FormatCurrency(amount float64) string {
	p := message.NewPrinter(language.Indonesian)
	return p.Sprintf("Rp %.0f", amount)
}

// FormatCurrencyString formats a numeric string to Rupiah currency string
func FormatCurrencyString(amountStr string) string {
	amount, _ := strconv.ParseFloat(amountStr, 64)
	return FormatCurrency(amount)
}

// ParseCurrencyString parses a formatted currency string back to a float64
func ParseCurrencyString(amountStr string) (float64, error) {
	amountStr = strings.ReplaceAll(amountStr, "Rp", "")
	amountStr = strings.ReplaceAll(amountStr, ".", "")
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amountStr = strings.TrimSpace(amountStr)
	return strconv.ParseFloat(amountStr, 64)
}

// ShowSuccessToast displays a non-blocking success message that auto-hides after 1 second
func ShowSuccessToast(title, message string, w fyne.Window) {
	info := dialog.NewInformation(title, message, w)
	info.Show()
	time.AfterFunc(1*time.Second, func() {
		info.Hide()
	})
}

// LoadPurchaseDisplayItems loads purchase items with full item information
func LoadPurchaseDisplayItems(s *state.Session, details []models.PurchaseDetail) []DisplayItem {
	items := make([]DisplayItem, len(details))

	for i, detail := range details {
		item, err := s.ItemRepo.GetByID(detail.ItemID)

		if err != nil {
			// Fallback if item not found
			items[i] = DisplayItem{
				Code:  "N/A",
				Name:  "Item tidak ditemukan",
				Qty:   fmt.Sprintf("%.0f", detail.Qty),
				Price: FormatCurrency(detail.PriceAmount),
				Total: FormatCurrency(detail.TotalAmount),
			}
		} else {
			items[i] = DisplayItem{
				Code:  item.Code,
				Name:  item.Name,
				Qty:   fmt.Sprintf("%.0f", detail.Qty),
				Price: FormatCurrency(detail.PriceAmount),
				Total: FormatCurrency(detail.TotalAmount),
			}
		}
	}

	return items
}

// LoadSellDisplayItems loads sell items with full item information
func LoadSellDisplayItems(s *state.Session, details []models.SellDetail) []DisplayItem {
	items := make([]DisplayItem, len(details))

	for i, detail := range details {
		item, err := s.ItemRepo.GetByID(detail.ItemID)

		if err != nil {
			// Fallback if item not found
			items[i] = DisplayItem{
				Code:  "N/A",
				Name:  "Item tidak ditemukan",
				Qty:   fmt.Sprintf("%.0f", detail.Qty),
				Price: FormatCurrency(detail.PriceAmount),
				Total: FormatCurrency(detail.TotalAmount),
			}
		} else {
			items[i] = DisplayItem{
				Code:  item.Code,
				Name:  item.Name,
				Qty:   fmt.Sprintf("%.0f", detail.Qty),
				Price: FormatCurrency(detail.PriceAmount),
				Total: FormatCurrency(detail.TotalAmount),
			}
		}
	}

	return items
}
