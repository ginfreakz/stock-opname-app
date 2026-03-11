package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID        uuid.UUID  `db:"id"`
	Username  string     `db:"username"`
	Name      string     `db:"name"`
	Password  string     `db:"password"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}
