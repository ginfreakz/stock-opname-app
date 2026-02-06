package ui

import (
	"fmt"
	"fyne-app/internal/models"
	"fyne-app/internal/state"
)

// Helper struct for displaying purchase/sell items with full details
type DisplayItem struct {
	Code  string
	Name  string
	Qty   string
	Price string
	Total string
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
				Price: fmt.Sprintf("%.0f", detail.PriceAmount),
				Total: fmt.Sprintf("%.0f", detail.TotalAmount),
			}
		} else {
			items[i] = DisplayItem{
				Code:  item.Code,
				Name:  item.Name,
				Qty:   fmt.Sprintf("%.0f", detail.Qty),
				Price: fmt.Sprintf("%.0f", detail.PriceAmount),
				Total: fmt.Sprintf("%.0f", detail.TotalAmount),
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
				Price: fmt.Sprintf("%.0f", detail.PriceAmount),
				Total: fmt.Sprintf("%.0f", detail.TotalAmount),
			}
		} else {
			items[i] = DisplayItem{
				Code:  item.Code,
				Name:  item.Name,
				Qty:   fmt.Sprintf("%.0f", detail.Qty),
				Price: fmt.Sprintf("%.0f", detail.PriceAmount),
				Total: fmt.Sprintf("%.0f", detail.TotalAmount),
			}
		}
	}
	
	return items
}