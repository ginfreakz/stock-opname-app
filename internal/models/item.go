package models

import (
	"time"
	"github.com/google/uuid"
)

type Item struct {
	ID         uuid.UUID  `db:"id"`
	Code       string     `db:"code"`
	Name       string     `db:"name"`
	Qty        float64    `db:"qty"`
	Price      float64    `db:"price"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
	CreatedBy  *uuid.UUID `db:"created_by"`
	UpdatedBy  *uuid.UUID `db:"updated_by"`
}