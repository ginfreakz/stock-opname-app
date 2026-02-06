package state

import (
	"fyne-app/internal/models"
	"fyne-app/internal/repository"
	"github.com/jmoiron/sqlx"
)

type Session struct {
	IsLoggedIn         bool
	Username           string
	User               *models.User
	DB                 *sqlx.DB
	UserRepo           *repository.UserRepository
	ItemRepo           *repository.ItemRepository
	PurchaseRepo       *repository.PurchaseRepository
	SellRepo           *repository.SellRepository
}

func NewSession(db *sqlx.DB) *Session {
	return &Session{
		DB:           db,
		UserRepo:     repository.NewUserRepository(db),
		ItemRepo:     repository.NewItemRepository(db),
		PurchaseRepo: repository.NewPurchaseRepository(db),
		SellRepo:     repository.NewSellRepository(db),
	}
}