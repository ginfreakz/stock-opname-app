package repository

import (
	"time"
	
	"fyne-app/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ItemRepository struct {
	db *sqlx.DB
}

func NewItemRepository(db *sqlx.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) GetAll() ([]models.Item, error) {
	var items []models.Item
	query := `SELECT id, code, name, qty, box_price, pack_price, rent_price, 
			  created_at, updated_at, deleted_at, created_by, updated_by 
			  FROM items 
			  WHERE deleted_at IS NULL 
			  ORDER BY name`
	
	err := r.db.Select(&items, query)
	return items, err
}

func (r *ItemRepository) GetByID(id uuid.UUID) (*models.Item, error) {
	var item models.Item
	query := `SELECT id, code, name, qty, box_price, pack_price, rent_price, 
			  created_at, updated_at, deleted_at, created_by, updated_by 
			  FROM items 
			  WHERE id = $1 AND deleted_at IS NULL`
	
	err := r.db.Get(&item, query, id)
	if err != nil {
		return nil, err
	}
	
	return &item, nil
}

func (r *ItemRepository) GetByCode(code string) (*models.Item, error) {
	var item models.Item
	query := `SELECT id, code, name, qty, box_price, pack_price, rent_price, 
			  created_at, updated_at, deleted_at, created_by, updated_by 
			  FROM items 
			  WHERE code = $1 AND deleted_at IS NULL`
	
	err := r.db.Get(&item, query, code)
	if err != nil {
		return nil, err
	}
	
	return &item, nil
}

func (r *ItemRepository) Create(item *models.Item) error {
	item.ID = uuid.New()
	item.CreatedAt = time.Now()
	
	query := `INSERT INTO items (id, code, name, qty, box_price, pack_price, rent_price, created_at, created_by) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	_, err := r.db.Exec(query, item.ID, item.Code, item.Name, item.Qty, 
		item.BoxPrice, item.PackPrice, item.RentPrice, item.CreatedAt, item.CreatedBy)
	return err
}

func (r *ItemRepository) Update(item *models.Item) error {
	now := time.Now()
	item.UpdatedAt = &now
	
	query := `UPDATE items 
			  SET name = $1, qty = $2, box_price = $3, pack_price = $4, 
			      rent_price = $5, updated_at = $6, updated_by = $7 
			  WHERE id = $8`
	
	_, err := r.db.Exec(query, item.Name, item.Qty, item.BoxPrice, 
		item.PackPrice, item.RentPrice, item.UpdatedAt, item.UpdatedBy, item.ID)
	return err
}

func (r *ItemRepository) Delete(id uuid.UUID, deletedBy uuid.UUID) error {
	now := time.Now()
	query := `UPDATE items 
			  SET deleted_at = $1, updated_by = $2, updated_at = $3 
			  WHERE id = $4`
	
	_, err := r.db.Exec(query, now, deletedBy, now, id)
	return err
}

func (r *ItemRepository) Search(keyword string) ([]models.Item, error) {
	var items []models.Item
	query := `SELECT id, code, name, qty, box_price, pack_price, rent_price, 
			  created_at, updated_at, deleted_at, created_by, updated_by 
			  FROM items 
			  WHERE deleted_at IS NULL 
			  AND (LOWER(code) LIKE LOWER($1) OR LOWER(name) LIKE LOWER($1))
			  ORDER BY name`
	
	searchPattern := "%" + keyword + "%"
	err := r.db.Select(&items, query, searchPattern)
	return items, err
}

func (r *ItemRepository) UpdateQty(tx *sqlx.Tx, itemID uuid.UUID, qtyChange float64) error {
	query := `UPDATE items SET qty = qty + $1 WHERE id = $2`
	
	var err error
	if tx != nil {
		_, err = tx.Exec(query, qtyChange, itemID)
	} else {
		_, err = r.db.Exec(query, qtyChange, itemID)
	}
	
	return err
}

func (r *ItemRepository) BeginTx() (*sqlx.Tx, error) {
	return r.db.Beginx()
}