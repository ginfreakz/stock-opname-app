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
	err = r.insertDetails(tx, &purchase.Header, purchase.Details)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PurchaseRepository) Update(purchase *models.PurchaseFull) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get old details to revert stock
	var oldDetails []models.PurchaseDetail
	err = tx.Select(&oldDetails, `SELECT item_id, qty FROM purchase_details WHERE header_id = $1`, purchase.Header.ID)
	if err != nil {
		return err
	}

	// 2. Revert stock and delete old mutations
	for _, d := range oldDetails {
		_, err = tx.Exec(`UPDATE items SET qty = qty - $1 WHERE id = $2`, d.Qty, d.ItemID)
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(`DELETE FROM stock_mutations WHERE model_id = $1 AND model_type = 'purchase'`, purchase.Header.ID)
	if err != nil {
		return err
	}

	// 3. Delete old details
	_, err = tx.Exec(`DELETE FROM purchase_details WHERE header_id = $1`, purchase.Header.ID)
	if err != nil {
		return err
	}

	// 4. Update header
	now := time.Now()
	purchase.Header.UpdatedAt = &now
	headerQuery := `UPDATE purchase_headers SET 
					purchase_invoice_num = $1, purchase_date = $2, supplier_name = $3, 
					updated_at = $4, updated_by = $5 
					WHERE id = $6`

	_, err = tx.Exec(headerQuery, purchase.Header.PurchaseInvoiceNum, purchase.Header.PurchaseDate,
		purchase.Header.SupplierName, purchase.Header.UpdatedAt,
		purchase.Header.UpdatedBy, purchase.Header.ID)
	if err != nil {
		return err
	}

	// 5. Insert new details and update stock
	err = r.insertDetails(tx, &purchase.Header, purchase.Details)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PurchaseRepository) insertDetails(tx *sqlx.Tx, header *models.PurchaseHeader, details []models.PurchaseDetail) error {
	for i := range details {
		detail := &details[i]
		detail.ID = uuid.New()
		detail.HeaderID = header.ID
		detail.CreatedAt = time.Now()

		detailQuery := `INSERT INTO purchase_details 
						(id, header_id, item_id, qty, price_amount, total_amount, created_at) 
						VALUES ($1, $2, $3, $4, $5, $6, $7)`

		_, err := tx.Exec(detailQuery, detail.ID, detail.HeaderID, detail.ItemID,
			detail.Qty, detail.PriceAmount, detail.TotalAmount, detail.CreatedAt)
		if err != nil {
			return err
		}

		// Update item quantity (add for purchases)
		_, err = tx.Exec(`UPDATE items SET qty = qty + $1 WHERE id = $2`, detail.Qty, detail.ItemID)
		if err != nil {
			return err
		}

		// Insert stock mutation
		mutationID := uuid.New()
		period := header.PurchaseDate.Format("2006-01")
		mutationQuery := `INSERT INTO stock_mutations 
						  (id, item_id, period, trx_date, qty, model_id, model_type, created_at) 
						  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

		_, err = tx.Exec(mutationQuery, mutationID, detail.ItemID, period,
			header.PurchaseDate, detail.Qty, header.ID, "purchase", time.Now())
		if err != nil {
			return err
		}
	}
	return nil
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
