package repository

import (
	"time"
	
	"fyne-app/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Authenticate(username, password string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, name, password, created_at, updated_at 
			  FROM users 
			  WHERE username = $1 AND password = $2`
	
	err := r.db.Get(&user, query, username, password)
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (r *UserRepository) Create(user *models.User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	
	query := `INSERT INTO users (id, username, name, password, created_at) 
			  VALUES ($1, $2, $3, $4, $5)`
	
	_, err := r.db.Exec(query, user.ID, user.Username, user.Name, user.Password, user.CreatedAt)
	return err
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, name, password, created_at, updated_at 
			  FROM users 
			  WHERE username = $1`
	
	err := r.db.Get(&user, query, username)
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}