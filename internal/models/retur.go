package models

import (
	"time"

	"github.com/google/uuid"
)

type ReturHeader struct {
	ID              uuid.UUID  `db:"id"`
	ReturInvoiceNum string     `db:"retur_invoice_num"`
	ReturDate       time.Time  `db:"retur_date"`
	SupplierName    string     `db:"supplier_name"`
	TotalAmount     float64    `db:"total_amount"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       *time.Time `db:"updated_at"`
	CreatedBy       *uuid.UUID `db:"created_by"`
	UpdatedBy       *uuid.UUID `db:"updated_by"`
}

type ReturDetail struct {
	ID          uuid.UUID  `db:"id"`
	HeaderID    uuid.UUID  `db:"header_id"`
	ItemID      uuid.UUID  `db:"item_id"`
	Qty         float64    `db:"qty"`
	PriceAmount float64    `db:"price_amount"`
	TotalAmount float64    `db:"total_amount"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
}

type ReturFull struct {
	Header  ReturHeader
	Details []ReturDetail
}
