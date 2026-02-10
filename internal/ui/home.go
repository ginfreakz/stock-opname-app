package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fyne-app/internal/state"
)

func showDeleteDataDialog(w fyne.Window, s *state.Session) {
	// Create confirmation message
	message := widget.NewLabel("Apakah Anda yakin ingin menghapus semua data transaksi yang lebih dari 3 tahun?")
	message.Wrapping = fyne.TextWrapWord

	warning := widget.NewLabel("⚠️ PERHATIAN: Tindakan ini tidak dapat dibatalkan!")
	warning.Wrapping = fyne.TextWrapWord
	warning.TextStyle = fyne.TextStyle{Bold: true}

	details := widget.NewLabel("Data yang akan dihapus:\n• Stock Mutations (> 3 tahun)\n• Purchase Headers & Details (> 3 tahun)\n• Sell Headers & Details (> 3 tahun)")
	details.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		message,
		widget.NewSeparator(),
		warning,
		widget.NewSeparator(),
		details,
	)

	// Create custom dialog
	var d dialog.Dialog

	confirmBtn := widget.NewButton("Ya, Hapus Data", func() {
		d.Hide()
		
		// Show progress dialog
		progressBar := widget.NewProgressBarInfinite()
		progressLabel := widget.NewLabel("Menghapus data...")
		progressContent := container.NewVBox(progressBar, progressLabel)
		
		progress := dialog.NewCustom("Menghapus Data", "", progressContent, w)
		progress.Show()

		// Channel to communicate result from goroutine
		type result struct {
			err        error
			successMsg string
		}
		resultChan := make(chan result, 1)

		// Perform deletion in background
		go func() {
			// Calculate date 3 years ago
			threeYearsAgo := time.Now().AddDate(-3, 0, 0)

			var deleteErr error
			var successMsg string

			// Start transaction
			tx, err := s.DB.Beginx()
			if err != nil {
				deleteErr = fmt.Errorf("Gagal memulai transaksi: %v", err)
			} else {
				defer tx.Rollback()

				// Delete stock mutations older than 3 years
				if deleteErr == nil {
					_, err = tx.Exec(`DELETE FROM stock_mutations WHERE trx_date < $1`, threeYearsAgo)
					if err != nil {
						deleteErr = fmt.Errorf("Gagal menghapus stock mutations: %v", err)
					}
				}

				// Delete purchase details for old purchases
				if deleteErr == nil {
					_, err = tx.Exec(`
						DELETE FROM purchase_details 
						WHERE header_id IN (
							SELECT id FROM purchase_headers WHERE purchase_date < $1
						)
					`, threeYearsAgo)
					if err != nil {
						deleteErr = fmt.Errorf("Gagal menghapus purchase details: %v", err)
					}
				}

				// Delete purchase headers older than 3 years
				if deleteErr == nil {
					_, err = tx.Exec(`DELETE FROM purchase_headers WHERE purchase_date < $1`, threeYearsAgo)
					if err != nil {
						deleteErr = fmt.Errorf("Gagal menghapus purchase headers: %v", err)
					}
				}

				// Delete sell details for old sells
				if deleteErr == nil {
					_, err = tx.Exec(`
						DELETE FROM sell_details 
						WHERE header_id IN (
							SELECT id FROM sell_headers WHERE sell_date < $1
						)
					`, threeYearsAgo)
					if err != nil {
						deleteErr = fmt.Errorf("Gagal menghapus sell details: %v", err)
					}
				}

				// Delete sell headers older than 3 years
				if deleteErr == nil {
					_, err = tx.Exec(`DELETE FROM sell_headers WHERE sell_date < $1`, threeYearsAgo)
					if err != nil {
						deleteErr = fmt.Errorf("Gagal menghapus sell headers: %v", err)
					}
				}

				// Commit transaction
				if deleteErr == nil {
					err = tx.Commit()
					if err != nil {
						deleteErr = fmt.Errorf("Gagal menyimpan perubahan: %v", err)
					} else {
						successMsg = fmt.Sprintf("Data transaksi sebelum tanggal %s berhasil dihapus!", 
							threeYearsAgo.Format("2006-01-02"))
					}
				}
			}

			// Send result through channel
			resultChan <- result{err: deleteErr, successMsg: successMsg}
		}()

		// Wait for result in a separate goroutine and update UI
		go func() {
			res := <-resultChan
			
			// ✅ Use fyne.Do to safely update UI from a background goroutine
			fyne.Do(func() {
				progress.Hide()
				
				if res.err != nil {
					dialog.ShowError(res.err, w)
				} else {
					dialog.ShowInformation("Berhasil", res.successMsg, w)
				}
			})
		}()
	})
	confirmBtn.Importance = widget.DangerImportance

	cancelBtn := widget.NewButton("Batal", func() {
		d.Hide()
	})

	buttons := container.NewGridWithColumns(2, cancelBtn, confirmBtn)

	finalContent := container.NewBorder(
		nil,
		buttons,
		nil,
		nil,
		content,
	)

	bg := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	dialogContent := container.NewMax(bg, container.NewPadded(finalContent))

	d = dialog.NewCustom("Konfirmasi Hapus Data", "", dialogContent, w)
	d.Resize(fyne.NewSize(500, 300))
	d.Show()
}

func HomePage(w fyne.Window, s *state.Session) fyne.CanvasObject {

	bg := canvas.NewImageFromFile("assets/bg-login.jpg")
	bg.FillMode = canvas.ImageFillStretch

	title := widget.NewLabelWithStyle(
		"Home",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	btnPenjualan := widget.NewButton("Penjualan Barang", func() {
		w.SetContent(PenjualanPage(w, s))
	})
	btnPembelian := widget.NewButton("Pembelian Barang", func() {
		w.SetContent(PembelianPage(w, s))
	})
	btnInventory := widget.NewButton("Inventory / Opname", func() {
		w.SetContent(InventoryPage(w, s))
	})
	btnHapus := widget.NewButton("Hapus Data", func() {
		showDeleteDataDialog(w, s)
	})
	btnHapus.Importance = widget.WarningImportance

	logout := widget.NewButton("Logout", func() {
		s.IsLoggedIn = false
		s.Username = ""
		s.User = nil
		w.SetContent(LoginPage(w, s))
	})
	logout.Importance = widget.DangerImportance

	separator := canvas.NewLine(color.Gray{Y: 120})
	separator.StrokeWidth = 2

	menu := container.NewVBox(
		title,
		btnPenjualan,
		btnPembelian,
		btnInventory,
		btnHapus,
		separator,
		logout,
	)

	card := widget.NewCard("", "", menu)

	cardContainer := container.NewGridWrap(
		fyne.NewSize(360, 300),
		card,
	)

	return container.NewMax(
		bg,
		container.NewCenter(cardContainer),
	)
}