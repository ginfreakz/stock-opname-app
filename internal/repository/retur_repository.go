package repository

import (
	"errors"
	"fmt"
	"time"

	"fyne-app/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ReturRepository struct {
	db *sqlx.DB
}

func NewReturRepository(db *sqlx.DB) *ReturRepository {
	return &ReturRepository{db: db}
}

// Create inserts a new ReturPembelian header and its details within a database transaction.
// It also updates the stock quantity of the associated items (qty = qty - retur_qty).
// Returns the newly created header ID, or an error if any step fails.
func (r *ReturRepository) Create(header *models.ReturHeader, details []models.ReturDetail) (uuid.UUID, error) {
	// Start transaction
	tx, err := r.db.Beginx()
	if err != nil {
		return uuid.Nil, err
	}

	// Defer rollback, it will be ignored if tx.Commit() succeeds
	defer tx.Rollback()

	// 1. Insert data ke tabel retur_headers
	header.ID = uuid.New()
	header.CreatedAt = time.Now()

	headerQuery := `INSERT INTO retur_headers 
		(id, retur_invoice_num, retur_date, supplier_name, total_amount, status, created_at, created_by) 
		VALUES ($1, $2, $3, $4, $5, 'ACTIVE', $6, $7)`

	_, err = tx.Exec(
		headerQuery,
		header.ID,
		header.ReturInvoiceNum,
		header.ReturDate,
		header.SupplierName,
		header.TotalAmount,
		header.CreatedAt,
		header.CreatedBy,
	)
	if err != nil {
		return uuid.Nil, err
	}

	// 2. Lakukan perulangan untuk INSERT retur_details
	for i := range details {
		detail := &details[i]
		detail.ID = uuid.New()
		detail.HeaderID = header.ID // menggunakan header_id dari langkah 1
		detail.CreatedAt = header.CreatedAt

		detailQuery := `INSERT INTO retur_details 
			(id, header_id, item_id, qty, price_amount, total_amount, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)`

		_, err = tx.Exec(
			detailQuery,
			detail.ID,
			detail.HeaderID,
			detail.ItemID,
			detail.Qty,
			detail.PriceAmount,
			detail.TotalAmount,
			detail.CreatedAt,
		)
		if err != nil {
			return uuid.Nil, err
		}

		// Validasi apakah qty retur melebihi qty stok saat ini
		if detail.Qty <= 0 {
			return uuid.Nil, errors.New("retur gagal: jumlah retur harus lebih besar dari 0")
		}

		var currentQty float64
		err = tx.Get(&currentQty, `SELECT qty FROM items WHERE id = $1`, detail.ItemID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("gagal mengecek stok barang: %v", err)
		}

		if detail.Qty > currentQty {
			return uuid.Nil, errors.New("retur gagal: jumlah retur melebihi stok yang tersedia saat ini")
		}

		// 3. Lakukan UPDATE ke tabel items untuk mengurangi qty
		// Logika: retur pembelian berarti barang dikembalikan ke supplier, sehingga stock gudang berkurang
		updateItemQuery := `UPDATE items SET qty = qty - $1 WHERE id = $2`
		_, err = tx.Exec(updateItemQuery, detail.Qty, detail.ItemID)
		if err != nil {
			return uuid.Nil, err
		}
	}

	// Jika semua berhasil, commit transaksi
	err = tx.Commit()
	if err != nil {
		return uuid.Nil, err
	}

	return header.ID, nil
}

// Update modifies an existing ReturPembelian header and its details within a database transaction.
// It reverts the old stock changes, deletes old details, updates the header, inserts new details,
// and applies new stock changes. Returns an error if any step fails.
func (r *ReturRepository) Update(header *models.ReturHeader, details []models.ReturDetail) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Dapatkan detail lama untuk mengembalikan stok
	var oldDetails []models.ReturDetail
	err = tx.Select(&oldDetails, `SELECT item_id, qty FROM retur_details WHERE header_id = $1`, header.ID)
	if err != nil {
		return fmt.Errorf("gagal mengambil detail lama: %v", err)
	}

	// 2. Kembalikan stok lama (karena retur memotong stok, kita kembalikan dengan menambah)
	for _, d := range oldDetails {
		_, err = tx.Exec(`UPDATE items SET qty = qty + $1 WHERE id = $2`, d.Qty, d.ItemID)
		if err != nil {
			return fmt.Errorf("gagal mengembalikan stok lama: %v", err)
		}
	}

	// 3. Hapus detail lama
	_, err = tx.Exec(`DELETE FROM retur_details WHERE header_id = $1`, header.ID)
	if err != nil {
		return fmt.Errorf("gagal menghapus detail lama: %v", err)
	}

	// 4. Update header
	headerQuery := `UPDATE retur_headers 
		SET retur_invoice_num = $1, retur_date = $2, supplier_name = $3, total_amount = $4, updated_at = $5, updated_by = $6 
		WHERE id = $7`
	_, err = tx.Exec(headerQuery, header.ReturInvoiceNum, header.ReturDate, header.SupplierName, header.TotalAmount, header.UpdatedAt, header.UpdatedBy, header.ID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui header: %v", err)
	}

	// 5. Insert detail baru dan update stok baru
	for i := range details {
		detail := &details[i]
		detail.ID = uuid.New()
		detail.HeaderID = header.ID
		detail.CreatedAt = *header.UpdatedAt // use updated_at as created_at for details since they are new

		detailQuery := `INSERT INTO retur_details 
			(id, header_id, item_id, qty, price_amount, total_amount, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)`

		_, err = tx.Exec(
			detailQuery,
			detail.ID,
			detail.HeaderID,
			detail.ItemID,
			detail.Qty,
			detail.PriceAmount,
			detail.TotalAmount,
			detail.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("gagal menyimpan detail baru: %v", err)
		}

		// Validasi stok baru
		if detail.Qty <= 0 {
			return errors.New("retur gagal: jumlah retur harus lebih besar dari 0")
		}

		var currentQty float64
		err = tx.Get(&currentQty, `SELECT qty FROM items WHERE id = $1`, detail.ItemID)
		if err != nil {
			return fmt.Errorf("gagal mengecek stok barang: %v", err)
		}

		if detail.Qty > currentQty {
			return errors.New("retur gagal: jumlah retur melebihi stok yang tersedia saat ini")
		}

		// Update stok baru (kurangi stok)
		updateItemQuery := `UPDATE items SET qty = qty - $1 WHERE id = $2`
		_, err = tx.Exec(updateItemQuery, detail.Qty, detail.ItemID)
		if err != nil {
			return fmt.Errorf("gagal memperbarui stok barang: %v", err)
		}
	}

	return tx.Commit()
}

// GetAll retrieves all Retur headers
func (r *ReturRepository) GetAll() ([]models.ReturHeader, error) {
	var headers []models.ReturHeader
	query := `SELECT id, retur_invoice_num, retur_date, supplier_name, total_amount, 
			  status, created_at, updated_at, created_by, updated_by 
			  FROM retur_headers 
			  ORDER BY retur_date DESC`

	err := r.db.Select(&headers, query)
	return headers, err
}

// GetByID retrieves a single Retur request including details
func (r *ReturRepository) GetByID(id uuid.UUID) (*models.ReturFull, error) {
	var retur models.ReturFull

	// Get header
	headerQuery := `SELECT id, retur_invoice_num, retur_date, supplier_name, total_amount,
					status, created_at, updated_at, created_by, updated_by 
					FROM retur_headers 
					WHERE id = $1`

	err := r.db.Get(&retur.Header, headerQuery, id)
	if err != nil {
		return nil, err
	}

	// Get details
	detailQuery := `SELECT id, header_id, item_id, qty, price_amount, total_amount, 
					created_at, updated_at 
					FROM retur_details 
					WHERE header_id = $1`

	err = r.db.Select(&retur.Details, detailQuery, id)
	if err != nil {
		return nil, err
	}

	return &retur, nil
}

// GetByInvoiceNum retrieves a Retur header by its invoice number
func (r *ReturRepository) GetByInvoiceNum(invoiceNum string) (*models.ReturHeader, error) {
	var header models.ReturHeader
	query := `SELECT id, retur_invoice_num, retur_date, supplier_name, total_amount, 
			  status, created_at, updated_at, created_by, updated_by 
			  FROM retur_headers 
			  WHERE retur_invoice_num = $1`

	err := r.db.Get(&header, query, invoiceNum)
	if err != nil {
		return nil, err // Return error (typically sql.ErrNoRows if not found)
	}
	return &header, nil
}

// Search retrieves Retur headers by invoice number or supplier name
func (r *ReturRepository) Search(keyword string) ([]models.ReturHeader, error) {
	var headers []models.ReturHeader
	query := `SELECT id, retur_invoice_num, retur_date, supplier_name, total_amount,
			  status, created_at, updated_at, created_by, updated_by 
			  FROM retur_headers 
			  WHERE LOWER(retur_invoice_num) LIKE LOWER($1) 
			  OR LOWER(supplier_name) LIKE LOWER($1)
			  ORDER BY retur_date DESC`

	searchPattern := "%" + keyword + "%"
	err := r.db.Select(&headers, query, searchPattern)
	return headers, err
}

// Void membatalkan transaksi Retur secara soft-delete
func (r *ReturRepository) Void(id uuid.UUID, updatedBy uuid.UUID) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Dapatkan detail item yang diretur
	var details []models.ReturDetail
	err = tx.Select(&details, `SELECT item_id, qty FROM retur_details WHERE header_id = $1`, id)
	if err != nil {
		return err
	}

	// 2. Kembalikan stok (tambah stok gudang karena batal kembalikan ke supplier)
	for _, d := range details {
		_, err = tx.Exec(`UPDATE items SET qty = qty + $1 WHERE id = $2`, d.Qty, d.ItemID)
		if err != nil {
			return err
		}
	}

	// 3. Catat mutasi pembatalan
	now := time.Now()
	period := now.Format("2006-01")
	for _, d := range details {
		mutationID := uuid.New()
		mutationQuery := `INSERT INTO stock_mutations 
						  (id, item_id, period, trx_date, qty, model_id, model_type, created_at) 
						  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

		// qty diset positif karena kita batal mengurangi stok retur (void retur)
		_, err = tx.Exec(mutationQuery, mutationID, d.ItemID, period, now, d.Qty, id, "void_retur", now)
		if err != nil {
			return err
		}
	}

	// 4. Update status header menjadi VOID
	_, err = tx.Exec(`UPDATE retur_headers SET status = 'VOID', updated_at = $1, updated_by = $2 WHERE id = $3 AND status != 'VOID'`, now, updatedBy, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
