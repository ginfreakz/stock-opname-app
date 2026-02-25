package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyne-app/internal/models"
	"fyne-app/internal/state"

	"github.com/google/uuid"
)

type PenjualanHeader struct {
	ID       uuid.UUID
	TglNota  string
	NoNota   string
	Customer string
}

type PenjualanItem struct {
	ItemID     uuid.UUID
	KodeBarang string
	NamaBarang string
	Qty        string
	Harga      string
	Total      string
}

type PenjualanFull struct {
	Header PenjualanHeader
	Items  []PenjualanItem
}

func showPenjualanDialog(w fyne.Window, s *state.Session, refreshCallback func(), initialFocus string, existingData *models.SellFull) {
	// Header form fields
	tglNota := widget.NewEntry()
	tglNota.SetText(time.Now().Format("2006-01-02"))
	tglNota.OnSubmitted = func(s string) {
		// No manual focus control for tglNota since it's in a border with calendarBtn
	}

	// Calendar button
	calendarBtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		dateEntry := widget.NewEntry()
		dateEntry.SetText(tglNota.Text)
		dateEntry.SetPlaceHolder("YYYY-MM-DD")

		dateDialog := dialog.NewCustomConfirm(
			"Pilih Tanggal",
			"OK",
			"Cancel",
			container.NewVBox(
				widget.NewLabel("Format: YYYY-MM-DD"),
				dateEntry,
			),
			func(ok bool) {
				if ok {
					tglNota.SetText(dateEntry.Text)
				}
			},
			w,
		)
		dateDialog.Show()
	})

	tglNotaContainer := container.NewBorder(nil, nil, nil, calendarBtn, tglNota)

	noNota := widget.NewEntry()
	customer := widget.NewEntry()

	// Item entry fields
	kodeBarang := widget.NewEntry()
	namaBarang := widget.NewEntry()
	qty := widget.NewEntry()

	// Focus flow
	noNota.OnSubmitted = func(s string) {
		w.Canvas().Focus(customer)
	}
	customer.OnSubmitted = func(s string) {
		w.Canvas().Focus(kodeBarang)
	}
	kodeBarang.OnSubmitted = func(s string) {
		w.Canvas().Focus(qty)
	}

	// Harga dropdown (combo select with arrow)
	// hargaOptions := []string{"Harga Dus", "Harga Pack", "Harga Rent"}
	// harga := widget.NewSelect(hargaOptions, func(value string) {
	// 	// Handle selection
	// })
	// harga.PlaceHolder = "Pilih Harga"

	// Store selected item for validation
	var selectedItem *models.Item
	var selectedItemIndex int = -1

	// Function to clear item form
	clearItemForm := func() {
		kodeBarang.SetText("")
		namaBarang.SetText("")
		qty.SetText("")
		selectedItem = nil
		selectedItemIndex = -1
	}

	// Kode barang lookup
	kodeBarang.OnChanged = func(code string) {
		if code == "" {
			namaBarang.SetText("")
			selectedItem = nil
			return
		}

		item, err := s.ItemRepo.GetByCode(code)
		if err == nil {
			namaBarang.SetText(item.Name)
			selectedItem = item
		} else {
			namaBarang.SetText("Item tidak ditemukan")
			selectedItem = nil
		}
	}

	// Items table data
	var items []PenjualanItem
	var itemsTable *widget.Table

	// Populate if existing
	if existingData != nil {
		tglNota.SetText(existingData.Header.SellDate.Format("2006-01-02"))
		noNota.SetText(existingData.Header.SellInvoiceNum)
		customer.SetText(existingData.Header.CustomerName)

		displayItems := LoadSellDisplayItems(s, existingData.Details)
		items = make([]PenjualanItem, len(displayItems))
		for i, v := range displayItems {
			items[i] = PenjualanItem{
				ItemID:     existingData.Details[i].ItemID,
				KodeBarang: v.Code,
				NamaBarang: v.Name,
				Qty:        v.Qty,
				Harga:      v.Price,
				Total:      v.Total,
			}
		}
	}

	// Function to refresh items table
	refreshItemsTable := func() {
		if itemsTable != nil {
			itemsTable.Refresh()
		}
	}

	// Buttons
	addItemBtn := widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), nil)
	addItemBtn.Importance = widget.HighImportance

	deleteItemBtn := widget.NewButtonWithIcon("Hapus Item", theme.DeleteIcon(), nil)
	deleteItemBtn.Importance = widget.DangerImportance

	clearBtn := widget.NewButtonWithIcon("Batal", theme.ContentClearIcon(), nil)

	updateButtonStates := func() {
		if selectedItemIndex >= 0 {
			addItemBtn.SetText("Update")
			addItemBtn.SetIcon(theme.DocumentSaveIcon())
			deleteItemBtn.Show()
			clearBtn.Show()
		} else {
			addItemBtn.SetText("Add")
			addItemBtn.SetIcon(theme.ContentAddIcon())
			deleteItemBtn.Hide()
			clearBtn.Hide()
		}
	}

	addItemBtn.OnTapped = func() {
		// Validate inputs
		if kodeBarang.Text == "" || selectedItem == nil || qty.Text == "" {
			dialog.ShowInformation("Error", "Kode barang dan qty harus diisi dan kode barang harus valid!", w)
			return
		}

		// Get price from selected item
		hargaVal := selectedItem.Price

		// Calculate total
		qtyVal, err := strconv.ParseFloat(qty.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Qty harus berupa angka!"), w)
			return
		}

		// Check stock availability
		// Special case: if updating, adding back the current qty to available stock for check
		availableStock := selectedItem.Qty
		if selectedItemIndex >= 0 {
			currentQty, _ := strconv.ParseFloat(items[selectedItemIndex].Qty, 64)
			availableStock += currentQty
		}

		if qtyVal > availableStock {
			dialog.ShowError(fmt.Errorf("Stok tidak mencukupi! Stok tersedia: %.0f", availableStock), w)
			return
		}

		total := qtyVal * hargaVal

		newItem := PenjualanItem{
			ItemID:     selectedItem.ID,
			KodeBarang: kodeBarang.Text,
			NamaBarang: namaBarang.Text,
			Qty:        qty.Text,
			Harga:      fmt.Sprintf("%.0f", hargaVal),
			Total:      fmt.Sprintf("%.0f", total),
		}

		if selectedItemIndex >= 0 {
			// Update existing
			items[selectedItemIndex] = newItem
		} else {
			// Add to items list
			items = append(items, newItem)
		}

		// Clear entry fields
		clearItemForm()
		updateButtonStates()

		// Refresh table
		refreshItemsTable()

		// Return focus to kodeBarang for next item
		w.Canvas().Focus(kodeBarang)
	}

	qty.OnSubmitted = func(s string) {
		addItemBtn.OnTapped()
	}

	deleteItemBtn.OnTapped = func() {
		if selectedItemIndex >= 0 {
			items = append(items[:selectedItemIndex], items[selectedItemIndex+1:]...)
			clearItemForm()
			updateButtonStates()
			refreshItemsTable()
		}
	}

	clearBtn.OnTapped = func() {
		clearItemForm()
		updateButtonStates()
	}

	updateButtonStates() // Initial state

	// Header form
	headerForm := widget.NewForm(
		widget.NewFormItem("Tgl. Nota", tglNotaContainer),
		widget.NewFormItem("No. Nota", noNota),
		widget.NewFormItem("Customer", customer),
	)
	headerFormSeparator := widget.NewSeparator()

	// Item entry form
	itemForm := widget.NewForm(
		widget.NewFormItem("Kode Barang", kodeBarang),
		widget.NewFormItem("Nama Barang", namaBarang),
		widget.NewFormItem("Qty", qty),
	)
	itemFormSeparator := widget.NewSeparator()
	itemFormButtons := container.NewHBox(addItemBtn, deleteItemBtn, clearBtn)

	if initialFocus == "item" {
		headerForm.Hide()
		headerFormSeparator.Hide()
	} else {
		// INS Mode: Hide Item entry
		itemForm.Hide()
		itemFormSeparator.Hide()
		itemFormButtons.Hide()
	}

	// Items table
	itemHeaders := []string{"Kode Barang", "Nama Barang", "QTY", "Harga", "Total", ""}
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	itemsTable = widget.NewTable(
		func() (int, int) {
			return len(items) + 1, len(itemHeaders)
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.White)
			text := canvas.NewText("", color.Black)
			text.Alignment = fyne.TextAlignCenter
			text.TextSize = 13
			btn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {})
			return container.NewMax(bg, text, btn)
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			cont := cell.(*fyne.Container)
			bg := cont.Objects[0].(*canvas.Rectangle)
			text := cont.Objects[1].(*canvas.Text)
			btn := cont.Objects[2].(*widget.Button)

			btn.Hide()

			if id.Row == 0 {
				bg.FillColor = headerBg
				text.Text = itemHeaders[id.Col]
				text.Color = color.White
				text.TextStyle = fyne.TextStyle{Bold: true}
				text.Alignment = fyne.TextAlignCenter
				text.Show()
				return
			}

			bg.FillColor = rowBg
			text.Color = color.Black
			text.TextStyle = fyne.TextStyle{}

			if id.Row-1 < len(items) {
				item := items[id.Row-1]
				switch id.Col {
				case 0:
					text.Text = item.KodeBarang
					text.Alignment = fyne.TextAlignLeading
					text.Show()
				case 1:
					text.Text = item.NamaBarang
					text.Alignment = fyne.TextAlignLeading
					text.Show()
				case 2:
					text.Text = item.Qty
					text.Alignment = fyne.TextAlignCenter
					text.Show()
				case 3:
					text.Text = item.Harga
					text.Alignment = fyne.TextAlignTrailing
					text.Show()
				case 4:
					text.Text = item.Total
					text.Alignment = fyne.TextAlignTrailing
					text.Show()
				case 5:
					text.Hide()
					btn.OnTapped = func() {
						rowIndex := id.Row - 1
						items = append(items[:rowIndex], items[rowIndex+1:]...)
						refreshItemsTable()
					}
					btn.Show()
				}
			} else {
				text.Text = ""
				text.Hide()
			}
			text.Refresh()
		},
	)

	itemsTable.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 && id.Row-1 < len(items) {
			selectedItemIndex = id.Row - 1
			item := items[selectedItemIndex]

			// Load item metadata to get the actual models.Item
			it, err := s.ItemRepo.GetByCode(item.KodeBarang)
			if err == nil {
				selectedItem = it
				kodeBarang.SetText(item.KodeBarang)
				namaBarang.SetText(item.NamaBarang)
				qty.SetText(item.Qty)
				updateButtonStates()
			}
		}
	}

	itemsTable.SetColumnWidth(0, 120) // Kode
	itemsTable.SetColumnWidth(1, 200) // Nama
	itemsTable.SetColumnWidth(2, 60)  // Qty
	itemsTable.SetColumnWidth(3, 120) // Harga
	itemsTable.SetColumnWidth(4, 120) // Total
	itemsTable.SetColumnWidth(5, 40)  // Delete

	// Dialog content
	var d dialog.Dialog

	submitBtn := widget.NewButton("Submit", func() {
		// Validate
		if tglNota.Text == "" || noNota.Text == "" || customer.Text == "" {
			dialog.ShowInformation("Error", "Header data harus diisi!", w)
			return
		}
		if initialFocus == "item" && len(items) == 0 {
			dialog.ShowInformation("Error", "Minimal 1 item harus ditambahkan!", w)
			return
		}

		// Parse date
		sellDate, err := time.Parse("2006-01-02", tglNota.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Format tanggal salah! Gunakan YYYY-MM-DD"), w)
			return
		}

		// Build sell model
		sell := &models.SellFull{
			Header: models.SellHeader{
				SellInvoiceNum: noNota.Text,
				SellDate:       sellDate,
				CustomerName:   customer.Text,
			},
			Details: make([]models.SellDetail, len(items)),
		}

		if existingData != nil {
			sell.Header.ID = existingData.Header.ID
			sell.Header.CreatedAt = existingData.Header.CreatedAt
			sell.Header.CreatedBy = existingData.Header.CreatedBy
			now := time.Now()
			sell.Header.UpdatedAt = &now
			sell.Header.UpdatedBy = &s.User.ID
		} else {
			sell.Header.CreatedBy = &s.User.ID
		}

		for i, item := range items {
			qtyVal, _ := strconv.ParseFloat(item.Qty, 64)
			hargaVal, _ := strconv.ParseFloat(item.Harga, 64)
			totalVal, _ := strconv.ParseFloat(item.Total, 64)

			sell.Details[i] = models.SellDetail{
				ItemID:      item.ItemID,
				Qty:         qtyVal,
				PriceAmount: hargaVal,
				TotalAmount: totalVal,
			}
		}

		// Save to database
		if existingData != nil {
			err = s.SellRepo.Update(sell)
		} else {
			err = s.SellRepo.Create(sell)
		}

		if err != nil {
			dialog.ShowError(fmt.Errorf("Gagal menyimpan data: %v", err), w)
			return
		}

		ShowSuccessToast("Success", "Data penjualan berhasil disimpan!", w)
		d.Hide()

		if refreshCallback != nil {
			refreshCallback()
		}
	})
	submitBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
	})
	cancelBtn.Importance = widget.DangerImportance

	buttons := container.NewGridWithColumns(2, cancelBtn, submitBtn)

	labelText := "Nota Baru"
	if initialFocus == "item" {
		labelText = "Isi Nota"
	}

	tableScroll := container.NewScroll(itemsTable)
	if initialFocus != "item" {
		tableScroll.Hide()
	}

	content := container.NewBorder(
		container.NewVBox(
			container.NewCenter(widget.NewLabelWithStyle("MENU PENJUALAN BARANG", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
			container.NewCenter(widget.NewLabelWithStyle(labelText, fyne.TextAlignCenter, fyne.TextStyle{})),
			widget.NewSeparator(),
			headerForm,
			headerFormSeparator,
			itemForm,
			itemFormButtons,
			itemFormSeparator,
		),
		buttons,
		nil,
		nil,
		tableScroll,
	)

	dialogContent := container.NewPadded(content)

	d = dialog.NewCustom("", "", dialogContent, w)
	d.Resize(fyne.NewSize(650, 600))
	d.Show()

	// Initial focus
	time.AfterFunc(100*time.Millisecond, func() {
		fyne.Do(func() {
			if initialFocus == "item" {
				w.Canvas().Focus(kodeBarang)
			} else {
				w.Canvas().Focus(tglNota)
			}
		})
	})
}

func showEditPenjualanDialog(w fyne.Window, s *state.Session, headerID uuid.UUID, refreshCallback func()) {
	sell, err := s.SellRepo.GetByID(headerID)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	showPenjualanDialog(w, s, refreshCallback, "item", sell)
}

func showViewPenjualanDialog(w fyne.Window, s *state.Session, headerID uuid.UUID) {
	// Load full sell data from database
	sell, err := s.SellRepo.GetByID(headerID)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Gagal memuat data: %v", err), w)
		return
	}

	// Read-only header info
	tglNota := widget.NewLabel(sell.Header.SellDate.Format("2006-01-02"))
	noNota := widget.NewLabel(sell.Header.SellInvoiceNum)
	customer := widget.NewLabel(sell.Header.CustomerName)

	headerInfo := container.NewVBox(
		container.NewGridWithColumns(2, canvas.NewText("Tgl. Nota", color.White), tglNota),
		container.NewGridWithColumns(2, canvas.NewText("No. Nota", color.White), noNota),
		container.NewGridWithColumns(2, canvas.NewText("Customer", color.White), customer),
	)

	// Convert details to display items using helper function
	displayItems := LoadSellDisplayItems(s, sell.Details)

	// Items table
	itemHeaders := []string{"Kode Barang", "Nama Barang", "QTY", "Harga", "Total"}
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	itemsTable := widget.NewTable(
		func() (int, int) {
			return len(displayItems) + 1, len(itemHeaders)
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			text := canvas.NewText("", color.Black)
			text.TextSize = 12
			return container.NewMax(bg, text)
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			cont := cell.(*fyne.Container)
			bg := cont.Objects[0].(*canvas.Rectangle)
			text := cont.Objects[1].(*canvas.Text)

			if id.Row == 0 {
				bg.FillColor = headerBg
				text.Text = itemHeaders[id.Col]
				text.Color = color.White
				text.TextSize = 13
				text.TextStyle = fyne.TextStyle{Bold: true}
				text.Alignment = fyne.TextAlignCenter
				text.Refresh()
				return
			}

			bg.FillColor = rowBg
			text.Color = color.Black
			text.TextStyle = fyne.TextStyle{}
			text.TextSize = 12

			item := displayItems[id.Row-1]
			switch id.Col {
			case 0:
				text.Text = item.Code
				text.Alignment = fyne.TextAlignLeading
			case 1:
				text.Text = item.Name
				text.Alignment = fyne.TextAlignLeading
			case 2:
				text.Text = item.Qty
				text.Alignment = fyne.TextAlignCenter
			case 3:
				text.Text = item.Price // This is already "Rp ..." if LoadSellDisplayItems was updated correctly
				text.Alignment = fyne.TextAlignTrailing
			case 4:
				text.Text = item.Total // This is already "Rp ..." if LoadSellDisplayItems was updated correctly
				text.Alignment = fyne.TextAlignTrailing
			}
			text.Refresh()
		},
	)

	itemsTable.SetColumnWidth(0, 110)
	itemsTable.SetColumnWidth(1, 170)
	itemsTable.SetColumnWidth(2, 70)
	itemsTable.SetColumnWidth(3, 110)
	itemsTable.SetColumnWidth(4, 110)

	var d dialog.Dialog

	closeBtn := widget.NewButton("Close", func() {
		d.Hide()
	})

	content := container.NewBorder(
		container.NewVBox(
			container.NewCenter(widget.NewLabelWithStyle("MENU PENJUALAN BARANG", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
			container.NewCenter(widget.NewLabelWithStyle("View data", fyne.TextAlignCenter, fyne.TextStyle{})),
			widget.NewSeparator(),
			headerInfo,
			widget.NewSeparator(),
		),
		container.NewCenter(closeBtn),
		nil,
		nil,
		container.NewScroll(itemsTable),
	)

	dialogContent := container.NewPadded(content)

	d = dialog.NewCustom("", "", dialogContent, w)
	d.Resize(fyne.NewSize(650, 500))
	d.Show()
}

func PenjualanPage(w fyne.Window, s *state.Session) fyne.CanvasObject {
	// Background
	bg := canvas.NewImageFromFile("assets/bg-login.jpg")
	bg.FillMode = canvas.ImageFillStretch

	// Header
	backBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		w.SetContent(HomePage(w, s))
	})
	backBtn.Importance = widget.LowImportance

	title := canvas.NewText("MENU PENJUALAN BARANG", color.White)
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 16

	search := widget.NewEntry()
	search.SetPlaceHolder("Search...")

	header := container.NewGridWithColumns(3, backBtn, container.NewCenter(title), container.NewMax(search))

	// Table headers
	headers := []string{"Tgl. Nota", "No. Nota", "Customer"}
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	var data []PenjualanHeader
	var selectedRow int = -1

	// Load data from database
	loadData := func(keyword string) {
		selectedRow = -1
		var headers []models.SellHeader
		var err error

		if keyword == "" {
			headers, err = s.SellRepo.GetAll()
		} else {
			headers, err = s.SellRepo.Search(keyword)
		}

		if err != nil {
			dialog.ShowError(fmt.Errorf("Gagal memuat data: %v", err), w)
			return
		}

		data = make([]PenjualanHeader, len(headers))
		for i, h := range headers {
			data[i] = PenjualanHeader{
				ID:       h.ID,
				TglNota:  h.SellDate.Format("2006-01-02"),
				NoNota:   h.SellInvoiceNum,
				Customer: h.CustomerName,
			}
		}
	}

	// Initial load
	loadData("")

	// Table
	table := widget.NewTable(
		func() (int, int) {
			return len(data) + 1, len(headers)
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			text := canvas.NewText("", color.Black)
			text.TextSize = 13
			return container.NewMax(bg, text)
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			cont := cell.(*fyne.Container)
			bg := cont.Objects[0].(*canvas.Rectangle)
			text := cont.Objects[1].(*canvas.Text)

			if id.Row == 0 {
				bg.FillColor = headerBg
				text.Text = headers[id.Col]
				text.Color = color.White
				text.TextSize = 14
				text.TextStyle = fyne.TextStyle{Bold: true}
				text.Alignment = fyne.TextAlignCenter
				return
			}

			if id.Row-1 == selectedRow {
				bg.FillColor = color.NRGBA{R: 100, G: 150, B: 255, A: 255}
				text.Color = color.White
			} else {
				bg.FillColor = rowBg
				text.Color = color.Black
			}

			text.TextStyle = fyne.TextStyle{}
			text.TextSize = 13

			if id.Row-1 < len(data) {
				item := data[id.Row-1]
				switch id.Col {
				case 0:
					text.Text = item.TglNota
					text.Alignment = fyne.TextAlignCenter
				case 1:
					text.Text = item.NoNota
					text.Alignment = fyne.TextAlignCenter
				case 2:
					text.Text = item.Customer
					text.Alignment = fyne.TextAlignLeading
				}
			}
		},
	)

	table.SetColumnWidth(0, 250)
	table.SetColumnWidth(1, 250)
	table.SetColumnWidth(2, 250)

	var focusWrapper *focusableTable

	safeFocus := func() {
		if focusWrapper != nil {
			fyne.Do(func() {
				w.Canvas().Focus(focusWrapper)
			})
		}
	}

	// Refresh function
	refreshTable := func() {
		loadData(search.Text)
		table.Refresh()
		safeFocus()
	}

	// Keyboard shortcuts handler
	handleKey := func(k *fyne.KeyEvent) {
		switch k.Name {
		case fyne.KeyInsert:
			showPenjualanDialog(w, s, refreshTable, "header", nil)
		case fyne.KeyE:
			if selectedRow >= 0 && selectedRow < len(data) {
				showEditPenjualanDialog(w, s, data[selectedRow].ID, refreshTable)
			} else {
				dialog.ShowInformation("Info", "Pilih nota terlebih dahulu!", w)
			}
		case fyne.KeyV:
			if selectedRow >= 0 && selectedRow < len(data) {
				showViewPenjualanDialog(w, s, data[selectedRow].ID)
			} else {
				dialog.ShowInformation("Info", "Pilih data terlebih dahulu!", w)
			}
		}
	}

	// Table selection
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 {
			selectedRow = id.Row - 1
			table.Refresh()
			time.AfterFunc(50*time.Millisecond, safeFocus)
		}
	}

	// ===== FOCUSABLE WRAPPER =====
	focusWrapper = newFocusableTable(table, handleKey)

	// Canvas-level fallback
	w.Canvas().SetOnTypedKey(handleKey)

	// Table wrapper (using focusWrapper instead of table)
	tableWrapper := container.NewCenter(
		container.NewGridWrap(
			fyne.NewSize(780, 400),
			focusWrapper,
		),
	)

	// Footer
	footer := canvas.NewText("[Insert] Nota Baru  [E] Isi Nota ", color.White)
	footer.TextStyle = fyne.TextStyle{Italic: true}

	// Content
	content := container.NewBorder(header, footer, nil, nil, tableWrapper)

	// Create a semi-transparent dark gray rectangle for the panel (matching login/home)
	rect := canvas.NewRectangle(color.NRGBA{R: 30, G: 30, B: 30, A: 180})
	rect.CornerRadius = 12
	rect.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 40} // Subtle white border
	rect.StrokeWidth = 1
	rect.SetMinSize(fyne.NewSize(980, 560))

	// Stack content on top of the background rectangle with padding
	panel := container.NewMax(
		rect,
		container.NewPadded(content),
	)

	// Center the panel in the window
	centeredPanel := container.NewCenter(panel)

	// Initial focus after 150ms
	time.AfterFunc(150*time.Millisecond, func() {
		fyne.Do(func() {
			safeFocus()
		})
	})

	return container.NewMax(
		bg,
		centeredPanel,
	)
}
