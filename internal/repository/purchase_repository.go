package repository

import (
	"time"
	
	"fyne-app/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PurchaseRepository struct {
	db *sqlx.DB
}

func NewPurchaseRepository(db *sqlx.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

func (r *PurchaseRepository) Create(purchase *models.PurchaseFull) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert header
	purchase.Header.ID = uuid.New()
	purchase.Header.CreatedAt = time.Now()
	
	headerQuery := `INSERT INTO purchase_headers 
					(id, purchase_invoice_num, purchase_date, supplier_name, created_at, created_by) 
					VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err = tx.Exec(headerQuery, purchase.Header.ID, purchase.Header.PurchaseInvoiceNum,
		purchase.Header.PurchaseDate, purchase.Header.SupplierName,
		purchase.Header.CreatedAt, purchase.Header.CreatedBy)
	if err != nil {
		return err
	}

	// Insert details and update stock
	for i := range purchase.Details {
		detail := &purchase.Details[i]
		detail.ID = uuid.New()
		detail.HeaderID = purchase.Header.ID
		detail.CreatedAt = time.Now()
		
		detailQuery := `INSERT INTO purchase_details 
						(id, header_id, item_id, qty, price_amount, total_amount, created_at) 
						VALUES ($1, $2, $3, $4, $5, $6, $7)`
		
		_, err = tx.Exec(detailQuery, detail.ID, detail.HeaderID, detail.ItemID,
			detail.Qty, detail.PriceAmount, detail.TotalAmount, detail.CreatedAt)
		if err != nil {
			return err
		}

		// Update item quantity
		_, err = tx.Exec(`UPDATE items SET qty = qty + $1 WHERE id = $2`, detail.Qty, detail.ItemID)
		if err != nil {
			return err
		}

		// Insert stock mutation
		mutationID := uuid.New()
		period := purchase.Header.PurchaseDate.Format("2006-01")
		mutationQuery := `INSERT INTO stock_mutations 
						  (id, item_id, period, trx_date, qty, model_id, model_type, created_at) 
						  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
		
		_, err = tx.Exec(mutationQuery, mutationID, detail.ItemID, period,
			purchase.Header.PurchaseDate, detail.Qty, purchase.Header.ID, "purchase", time.Now())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PurchaseRepository) GetAll() ([]models.PurchaseHeader, error) {
	var headers []models.PurchaseHeader
	query := `SELECT id, purchase_invoice_num, purchase_date, supplier_name, 
			  created_at, updated_at, created_by, updated_by 
			  FROM purchase_headers 
			  ORDER BY purchase_date DESC`
	
	err := r.db.Select(&headers, query)
	return headers, err
}

func (r *PurchaseRepository) GetByID(id uuid.UUID) (*models.PurchaseFull, error) {
	var purchase models.PurchaseFull
	
	// Get header
	headerQuery := `SELECT id, purchase_invoice_num, purchase_date, supplier_name, 
					created_at, updated_at, created_by, updated_by 
					FROM purchase_headers 
					WHERE id = $1`
	
	err := r.db.Get(&purchase.Header, headerQuery, id)
	if err != nil {
		return nil, err
	}

	// Get details
	detailQuery := `SELECT id, header_id, item_id, qty, price_amount, total_amount, 
					created_at, updated_at 
					FROM purchase_details 
					WHERE header_id = $1`
	
	err = r.db.Select(&purchase.Details, detailQuery, id)
	if err != nil {
		return nil, err
	}

	return &purchase, nil
}

func (r *PurchaseRepository) Search(keyword string) ([]models.PurchaseHeader, error) {
	var headers []models.PurchaseHeader
	query := `SELECT id, purchase_invoice_num, purchase_date, supplier_name, 
			  created_at, updated_at, created_by, updated_by 
			  FROM purchase_headers 
			  WHERE LOWER(purchase_invoice_num) LIKE LOWER($1) 
			  OR LOWER(supplier_name) LIKE LOWER($1)
			  ORDER BY purchase_date DESC`
	
	searchPattern := "%" + keyword + "%"
	err := r.db.Select(&headers, query, searchPattern)
	return headers, err
}