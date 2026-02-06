package models

import (
	"time"
	"github.com/google/uuid"
)

type SellHeader struct {
	ID             uuid.UUID  `db:"id"`
	SellInvoiceNum string     `db:"sell_invoice_num"`
	SellDate       time.Time  `db:"sell_date"`
	CustomerName   string     `db:"customer_name"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      *time.Time `db:"updated_at"`
	CreatedBy      *uuid.UUID `db:"created_by"`
	UpdatedBy      *uuid.UUID `db:"updated_by"`
}

type SellDetail struct {
	ID          uuid.UUID  `db:"id"`
	HeaderID    uuid.UUID  `db:"header_id"`
	ItemID      uuid.UUID  `db:"item_id"`
	Qty         float64    `db:"qty"`
	PriceAmount float64    `db:"price_amount"`
	TotalAmount float64    `db:"total_amount"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
}

type SellFull struct {
	Header  SellHeader
	Details []SellDetail
}