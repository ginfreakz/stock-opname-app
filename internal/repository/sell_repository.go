package repository

import (
	"time"

	"fyne-app/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SellRepository struct {
	db *sqlx.DB
}

func NewSellRepository(db *sqlx.DB) *SellRepository {
	return &SellRepository{db: db}
}

func (r *SellRepository) Create(sell *models.SellFull) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert header
	sell.Header.ID = uuid.New()
	sell.Header.CreatedAt = time.Now()

	headerQuery := `INSERT INTO sell_headers 
					(id, sell_invoice_num, sell_date, customer_name, created_at, created_by) 
					VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = tx.Exec(headerQuery, sell.Header.ID, sell.Header.SellInvoiceNum,
		sell.Header.SellDate, sell.Header.CustomerName,
		sell.Header.CreatedAt, sell.Header.CreatedBy)
	if err != nil {
		return err
	}

	// Insert details and update stock
	return r.insertDetails(tx, &sell.Header, sell.Details)
}

func (r *SellRepository) Update(sell *models.SellFull) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get old details to revert stock
	var oldDetails []models.SellDetail
	err = tx.Select(&oldDetails, `SELECT item_id, qty FROM sell_details WHERE header_id = $1`, sell.Header.ID)
	if err != nil {
		return err
	}

	// 2. Revert stock and delete old mutations
	for _, d := range oldDetails {
		_, err = tx.Exec(`UPDATE items SET qty = qty + $1 WHERE id = $2`, d.Qty, d.ItemID)
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(`DELETE FROM stock_mutations WHERE model_id = $1 AND model_type = 'sell'`, sell.Header.ID)
	if err != nil {
		return err
	}

	// 3. Delete old details
	_, err = tx.Exec(`DELETE FROM sell_details WHERE header_id = $1`, sell.Header.ID)
	if err != nil {
		return err
	}

	// 4. Update header
	now := time.Now()
	sell.Header.UpdatedAt = &now
	headerQuery := `UPDATE sell_headers SET 
					sell_invoice_num = $1, sell_date = $2, customer_name = $3, 
					updated_at = $4, updated_by = $5 
					WHERE id = $6`

	_, err = tx.Exec(headerQuery, sell.Header.SellInvoiceNum, sell.Header.SellDate,
		sell.Header.CustomerName, sell.Header.UpdatedAt,
		sell.Header.UpdatedBy, sell.Header.ID)
	if err != nil {
		return err
	}

	// 5. Insert new details and update stock
	err = r.insertDetails(tx, &sell.Header, sell.Details)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SellRepository) insertDetails(tx *sqlx.Tx, header *models.SellHeader, details []models.SellDetail) error {
	for i := range details {
		detail := &details[i]
		detail.ID = uuid.New()
		detail.HeaderID = header.ID
		detail.CreatedAt = time.Now()

		detailQuery := `INSERT INTO sell_details 
						(id, header_id, item_id, qty, price_amount, total_amount, created_at) 
						VALUES ($1, $2, $3, $4, $5, $6, $7)`

		_, err := tx.Exec(detailQuery, detail.ID, detail.HeaderID, detail.ItemID,
			detail.Qty, detail.PriceAmount, detail.TotalAmount, detail.CreatedAt)
		if err != nil {
			return err
		}

		// Update item quantity (subtract for sales)
		_, err = tx.Exec(`UPDATE items SET qty = qty - $1 WHERE id = $2`, detail.Qty, detail.ItemID)
		if err != nil {
			return err
		}

		// Insert stock mutation (negative quantity for sales)
		mutationID := uuid.New()
		period := header.SellDate.Format("2006-01")
		mutationQuery := `INSERT INTO stock_mutations 
						  (id, item_id, period, trx_date, qty, model_id, model_type, created_at) 
						  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

		_, err = tx.Exec(mutationQuery, mutationID, detail.ItemID, period,
			header.SellDate, -detail.Qty, header.ID, "sell", time.Now())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *SellRepository) GetAll() ([]models.SellHeader, error) {
	var headers []models.SellHeader
	query := `SELECT id, sell_invoice_num, sell_date, customer_name, 
			  created_at, updated_at, created_by, updated_by 
			  FROM sell_headers 
			  ORDER BY sell_date DESC`

	err := r.db.Select(&headers, query)
	return headers, err
}

func (r *SellRepository) GetByID(id uuid.UUID) (*models.SellFull, error) {
	var sell models.SellFull

	// Get header
	headerQuery := `SELECT id, sell_invoice_num, sell_date, customer_name, 
					created_at, updated_at, created_by, updated_by 
					FROM sell_headers 
					WHERE id = $1`

	err := r.db.Get(&sell.Header, headerQuery, id)
	if err != nil {
		return nil, err
	}

	// Get details
	detailQuery := `SELECT id, header_id, item_id, qty, price_amount, total_amount, 
					created_at, updated_at 
					FROM sell_details 
					WHERE header_id = $1`

	err = r.db.Select(&sell.Details, detailQuery, id)
	if err != nil {
		return nil, err
	}

	return &sell, nil
}

func (r *SellRepository) Search(keyword string) ([]models.SellHeader, error) {
	var headers []models.SellHeader
	query := `SELECT id, sell_invoice_num, sell_date, customer_name, 
			  created_at, updated_at, created_by, updated_by 
			  FROM sell_headers 
			  WHERE LOWER(sell_invoice_num) LIKE LOWER($1) 
			  OR LOWER(customer_name) LIKE LOWER($1)
			  ORDER BY sell_date DESC`

	searchPattern := "%" + keyword + "%"
	err := r.db.Select(&headers, query, searchPattern)
	return headers, err
}
