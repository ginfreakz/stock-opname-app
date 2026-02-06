package models

import (
	"time"
	"github.com/google/uuid"
)

type StockMutation struct {
	ID        uuid.UUID `db:"id"`
	ItemID    uuid.UUID `db:"item_id"`
	Period    string    `db:"period"`
	TrxDate   time.Time `db:"trx_date"`
	Qty       float64   `db:"qty"`
	ModelID   uuid.UUID `db:"model_id"`
	ModelType string    `db:"model_type"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}