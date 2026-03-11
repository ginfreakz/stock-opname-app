package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
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
	Total    string
	Status   string
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

func showPenjualanDialog(w fyne.Window, s *state.Session, refreshCallback func(), existingData *models.SellFull) {
	// Header form fields
	tglNota := widget.NewEntry()
	tglNota.SetText(time.Now().Format("2006-01-02"))
	// disable the entry so user can't type or select
	tglNota.Disable()

	// Calendar button with calendar icon only
	calendarBtn := widget.NewButtonWithIcon("", theme.CalendarIcon(), func() {
		ShowDatePickerDialog(w, tglNota.Text, func(selectedDate string) {
			tglNota.SetText(selectedDate)
		})
	})
	calendarBtn.Importance = widget.LowImportance

	tglNotaContainer := container.NewBorder(nil, nil, nil, calendarBtn, tglNota)

	noNota := widget.NewEntry()
	customer := widget.NewEntry()

	// Item entry fields
	kodeBarang := widget.NewEntry()

	var itemOptions []string
	allItems, _ := s.ItemRepo.GetAll()
	for _, it := range allItems {
		itemOptions = append(itemOptions, fmt.Sprintf("%s - %s", it.Code, it.Name))
	}

	namaBarang := widget.NewSelectEntry(itemOptions)
	namaBarang.PlaceHolder = "Pilih Barang..."

	qty := widget.NewEntry()

	stockInfo := canvas.NewText("", color.NRGBA{R: 128, G: 128, B: 128, A: 255})
	stockInfo.TextSize = 12
	stockWarning := canvas.NewText("", color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	stockWarning.TextSize = 12
	stockWarning.TextStyle = fyne.TextStyle{Italic: true}

	qtyContainer := container.NewVBox(
		qty,
		container.NewHBox(stockInfo, layout.NewSpacer(), stockWarning),
	)

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
		stockInfo.Text = ""
		stockWarning.Text = ""
		stockInfo.Refresh()
		stockWarning.Refresh()
	}

	isSyncing := false

	namaBarang.OnChanged = func(value string) {
		if isSyncing {
			return
		}

		parts := strings.SplitN(value, " - ", 2)
		if len(parts) == 2 {
			isSyncing = true
			kodeBarang.SetText(parts[0])
			isSyncing = false

			item, err := s.ItemRepo.GetByCode(parts[0])
			if err == nil {
				selectedItem = item
				stockInfo.Text = fmt.Sprintf("Stok tersedia: %.0f", item.Qty)
			} else {
				selectedItem = nil
				stockInfo.Text = ""
			}
			stockWarning.Text = ""
			stockInfo.Refresh()
			stockWarning.Refresh()
		} else {
			isSyncing = true
			kodeBarang.SetText("")
			isSyncing = false
			selectedItem = nil
			stockInfo.Text = ""
			stockWarning.Text = ""
			stockInfo.Refresh()
			stockWarning.Refresh()
		}
	}

	// Items table data
	var items []PenjualanItem
	var itemsTable *widget.Table

	// Kode barang lookup
	kodeBarang.OnChanged = func(code string) {
		if isSyncing {
			return
		}

		var filtered []string
		if code == "" {
			filtered = itemOptions
		} else {
			lowerCode := strings.ToLower(code)
			for _, opt := range itemOptions {
				parts := strings.SplitN(opt, " - ", 2)
				if len(parts) == 2 {
					if strings.Contains(strings.ToLower(parts[0]), lowerCode) {
						filtered = append(filtered, opt)
					}
				}
			}
		}

		isSyncing = true
		namaBarang.SetOptions(filtered)
		isSyncing = false

		if code == "" {
			isSyncing = true
			namaBarang.SetText("")
			isSyncing = false
			selectedItem = nil
			return
		}

		item, err := s.ItemRepo.GetByCode(code)
		if err == nil {
			isSyncing = true
			namaBarang.SetText(fmt.Sprintf("%s - %s", item.Code, item.Name))
			isSyncing = false
			selectedItem = item

			var currentInCart float64
			for i, row := range items {
				if row.ItemID == selectedItem.ID && i != selectedItemIndex {
					qtyVal, _ := strconv.ParseFloat(row.Qty, 64)
					currentInCart += qtyVal
				}
			}
			stockInfo.Text = fmt.Sprintf("Stok tersedia: %.0f", item.Qty - currentInCart)
		} else {
			isSyncing = true
			namaBarang.SetText("")
			isSyncing = false
			selectedItem = nil
			stockInfo.Text = ""
		}
		stockWarning.Text = ""
		stockInfo.Refresh()
		stockWarning.Refresh()
	}

	qty.OnChanged = func(val string) {
		if selectedItem != nil {
			var currentInCart float64
			for i, item := range items {
				if item.ItemID == selectedItem.ID && i != selectedItemIndex {
					qtyVal, _ := strconv.ParseFloat(item.Qty, 64)
					currentInCart += qtyVal
				}
			}
			availableStock := selectedItem.Qty - currentInCart
			
			// Selalu update label info ketersediaan stok secara real-time berdasarkan isi cart
			stockInfo.Text = fmt.Sprintf("Stok tersedia: %.0f", availableStock)
			stockInfo.Refresh()

			v, err := strconv.ParseFloat(val, 64)
			if err == nil {
				if v > availableStock { // For Penjualan, stock must be sufficient
					stockWarning.Text = "Stok kurang!"
				} else if v <= 0 {
					stockWarning.Text = "Qty invalid!"
				} else {
					stockWarning.Text = ""
				}
			} else {
				stockWarning.Text = ""
			}
			stockWarning.Refresh()
		}
	}

	// Items table data variables moved above qty.OnChanged

	updateStockInfo := func() {
		if selectedItem != nil {
			var currentInCart float64
			for i, row := range items {
				if row.ItemID == selectedItem.ID && i != selectedItemIndex {
					qtyVal, _ := strconv.ParseFloat(row.Qty, 64)
					currentInCart += qtyVal
				}
			}
			stockInfo.Text = fmt.Sprintf("Stok tersedia: %.0f", selectedItem.Qty - currentInCart)
			stockInfo.Refresh()
		}
	}

	totalLabel := canvas.NewText("Total : Rp 0", color.Black)
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}
	totalLabel.Alignment = fyne.TextAlignTrailing

	recalculateTotal := func() float64 {
		var sum float64
		for _, item := range items {
			val, _ := ParseCurrencyString(item.Total)
			sum += val
		}
		totalLabel.Text = "Total : " + FormatCurrency(sum)
		totalLabel.Refresh()
		return sum
	}

	// Function to refresh items table
	refreshItemsTable := func() {
		if itemsTable != nil {
			itemsTable.Refresh()
			recalculateTotal()
			updateStockInfo()
		}
	}

	// Header form
	headerForm := widget.NewForm(
		widget.NewFormItem("Tgl. Nota", tglNotaContainer),
		widget.NewFormItem("No. Nota", noNota),
		widget.NewFormItem("Customer", customer),
	)
	headerFormSeparator := widget.NewSeparator()

	// Populate if existing
	if existingData != nil {
		formattedDate := existingData.Header.SellDate.Format("2006-01-02")
		tglNota.SetText(formattedDate)
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

		// Lock header inputs
		tglNota.Disable()
		noNota.Disable()
		customer.Disable()
		calendarBtn.Disable()

		recalculateTotal()
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
		if qtyVal <= 0 {
			dialog.ShowInformation("Error", "Qty harus lebih besar dari 0!", w)
			return
		}

		// Check stock availability (Penjualan / Jual = kurang stock)
		// Harus cek ketersediaan dikurang qty yang sudah masuk keranjang di transaksi ini
		var currentInCart float64
		for i, item := range items {
			if item.ItemID == selectedItem.ID && i != selectedItemIndex {
				q, _ := strconv.ParseFloat(item.Qty, 64)
				currentInCart += q
			}
		}

		availableStock := selectedItem.Qty - currentInCart

		if qtyVal > availableStock {
			dialog.ShowError(fmt.Errorf("Stok tidak mencukupi! Sisa stok yang bisa ditarik: %.0f", availableStock), w)
			return
		}

		total := qtyVal * hargaVal

		newItem := PenjualanItem{
			ItemID:     selectedItem.ID,
			KodeBarang: kodeBarang.Text,
			NamaBarang: namaBarang.Text,
			Qty:        qty.Text,
			Harga:      FormatCurrency(hargaVal),
			Total:      FormatCurrency(total),
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

	// Item entry form
	itemForm := widget.NewForm(
		widget.NewFormItem("Kode Barang", kodeBarang),
		widget.NewFormItem("Nama Barang", namaBarang),
		widget.NewFormItem("Qty", qtyContainer),
	)
	itemFormSeparator := widget.NewSeparator()
	itemFormButtons := container.NewHBox(addItemBtn, deleteItemBtn, clearBtn)

	// Merged dialog: always show both forms
	// Removed initialFocus logic

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
					if existingData != nil {
						btn.Hide()
						text.Text = ""
						text.Hide()
					} else {
						text.Hide()
						btn.OnTapped = func() {
							rowIndex := id.Row - 1
							items = append(items[:rowIndex], items[rowIndex+1:]...)
							refreshItemsTable()
						}
						btn.Show()
					}
				}
			} else {
				text.Text = ""
				text.Hide()
			}
			text.Refresh()
		},
	)

	itemsTable.OnSelected = func(id widget.TableCellID) {
		if existingData != nil { // Disable selection for existingData
			return
		}
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

	if existingData != nil {
		itemsTable.SetColumnWidth(4, 160) // Absorb Delete button width (120+40)
		itemsTable.SetColumnWidth(5, 0)   // Hide edit button for readonly mode
	} else {
		itemsTable.SetColumnWidth(4, 120) // Total
		itemsTable.SetColumnWidth(5, 40)  // Delete/Edit
	}

	// Dialog content
	var d dialog.Dialog

	submitBtn := widget.NewButton("Submit", func() {
		// Validate
		if tglNota.Text == "" || noNota.Text == "" || customer.Text == "" {
			dialog.ShowInformation("Error", "Header data harus diisi!", w)
			return
		}

		if existingData == nil {
			// Validasi No Nota duplikat
			existing, _ := s.SellRepo.GetByInvoiceNum(noNota.Text)
			if existing != nil {
				dialog.ShowInformation("Error", "No. Nota sudah terdaftar, silakan gunakan nomor lain!", w)
				return
			}
		}

		// Validate item count
		if len(items) == 0 {
			dialog.ShowInformation("Error", "Minimal 1 item harus ditambahkan!", w)
			return
		}

		// Parse date
		sellDate, err := time.Parse("2006-01-02", tglNota.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Format tanggal salah! Gunakan YYYY-MM-DD"), w)
			return
		}

		grandTotal := recalculateTotal()

		// Build sell model
		sell := &models.SellFull{
			Header: models.SellHeader{
				SellInvoiceNum: noNota.Text,
				SellDate:       sellDate,
				CustomerName:   customer.Text,
				TotalAmount:    grandTotal,
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
			hargaVal, _ := ParseCurrencyString(item.Harga)
			totalVal, _ := ParseCurrencyString(item.Total)

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

		// if date changed from original, mark blue
		// (UI color feedback removed)

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

	var buttons *fyne.Container

	if existingData != nil {
		cancelBtn.SetText("Tutup")
		cancelBtn.Importance = widget.HighImportance
		buttons = container.NewGridWithColumns(1, cancelBtn)
		submitBtn.Hide() // Hide submit button for existingData
	} else {
		buttons = container.NewGridWithColumns(2, cancelBtn, submitBtn)
	}

	labelText := "Nota Baru"
	if existingData != nil {
		labelText = "Detail Nota"

		// Hide edit elements completely
		kodeBarang.Disable()
		namaBarang.Disable()
		qty.Disable()
		addItemBtn.Hide()
		deleteItemBtn.Hide()
		clearBtn.Hide()
		kodeBarang.OnChanged = nil // Disable submit handlers
		namaBarang.OnChanged = nil
		qty.OnChanged = nil
		qty.OnSubmitted = nil
		addItemBtn.OnTapped = nil
		deleteItemBtn.OnTapped = nil
		clearBtn.OnTapped = nil
	}

	tableSection := container.NewBorder(
		nil,
		container.NewVBox(
			widget.NewSeparator(),
			container.NewHBox(
				layout.NewSpacer(),
				totalLabel,
			),
		),
		nil,
		nil,
		func() fyne.CanvasObject {
			scroll := container.NewScroll(itemsTable)
			scroll.SetMinSize(fyne.NewSize(0, 150))
			return scroll
		}(),
	)

	// tableSection is always shown

	topContent := container.NewVBox(
		container.NewCenter(widget.NewLabelWithStyle(
			"MENU PENJUALAN BARANG",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		)),
		container.NewCenter(widget.NewLabelWithStyle(labelText, fyne.TextAlignCenter, fyne.TextStyle{})),
		widget.NewSeparator(),
	)

	if existingData != nil {
		topContent.Add(headerForm)
		topContent.Add(headerFormSeparator)
		// Hide itemForm and itemFormButtons for existing data
	} else {
		topContent.Add(headerForm)
		topContent.Add(headerFormSeparator)
		topContent.Add(itemForm)
		topContent.Add(itemFormButtons)
		topContent.Add(itemFormSeparator)
	}

	content := container.NewBorder(
		topContent,
		buttons,
		nil,
		nil,
		tableSection,
	)

	dialogContent := container.NewPadded(content)

	d = dialog.NewCustom("", "", dialogContent, w)
	d.Resize(fyne.NewSize(750, 600))
	d.Show()

	// Initial focus
	time.AfterFunc(100*time.Millisecond, func() {
		fyne.Do(func() {
			w.Canvas().Focus(kodeBarang)
		})
	})
}

// (Edit dialog removed to unify Add flow)

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

	totalLabel := canvas.NewText(
		"Total : "+FormatCurrency(sell.Header.TotalAmount),
		color.Black,
	)
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}
	totalLabel.Alignment = fyne.TextAlignTrailing

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

	tableSection := container.NewBorder(
		nil,
		container.NewVBox(
			widget.NewSeparator(),
			container.NewHBox(
				layout.NewSpacer(),
				totalLabel,
			),
		),
		nil,
		nil,
		func() fyne.CanvasObject {
			scroll := container.NewScroll(itemsTable)
			scroll.SetMinSize(fyne.NewSize(0, 200))
			return scroll
		}(),
	)

	content := container.NewBorder(
		container.NewVBox(
			container.NewCenter(widget.NewLabelWithStyle(
				"MENU PENJUALAN BARANG",
				fyne.TextAlignCenter,
				fyne.TextStyle{Bold: true},
			)),
			container.NewCenter(widget.NewLabelWithStyle("View data", fyne.TextAlignCenter, fyne.TextStyle{})),
			widget.NewSeparator(),
			headerInfo,
			widget.NewSeparator(),
		),
		container.NewCenter(closeBtn),
		nil,
		nil,
		tableSection,
	)

	dialogContent := container.NewPadded(content)

	d = dialog.NewCustom("", "", dialogContent, w)
	d.Resize(fyne.NewSize(700, 500))
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

	title := canvas.NewText("MENU PENJUALAN BARANG", color.White)
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 16

	search := widget.NewEntry()
	search.SetPlaceHolder("Search No. Nota or Customer...")

	header := container.NewGridWithColumns(3, backBtn, container.NewCenter(title), container.NewMax(search))

	// Table headers
	headers := []string{"Tgl. Nota", "No. Nota", "Customer", "Total", "Aksi"}
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
				Total:    FormatCurrency(h.TotalAmount),
				Status:   h.Status,
			}
		}
	}

	// Initial load
	loadData("")

	// After initial load we will bind the search field once the table is created

	// Table
	table := widget.NewTable(
		func() (int, int) {
			return len(data) + 1, len(headers)
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			text := canvas.NewText("", color.Black)
			text.TextSize = 13
			text.Alignment = fyne.TextAlignCenter
			btn := widget.NewButtonWithIcon("", theme.DocumentPrintIcon(), nil)
			btn.Importance = widget.LowImportance
			return container.NewMax(bg, text, container.NewCenter(btn))
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			cont := cell.(*fyne.Container)
			bg := cont.Objects[0].(*canvas.Rectangle)
			text := cont.Objects[1].(*canvas.Text)
			btnCont := cont.Objects[2].(*fyne.Container)
			btn := btnCont.Objects[0].(*widget.Button)

			btn.Hide()
			text.Show()

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

				// Red text indicating VOID for all columns
				if item.Status == "VOID" {
					if id.Row-1 != selectedRow {
						text.Color = color.NRGBA{R: 220, G: 50, B: 50, A: 255}
					}
				}

				switch id.Col {
				case 0:
					text.Text = item.TglNota
					text.Alignment = fyne.TextAlignCenter
				case 1:
					if item.Status == "VOID" {
						text.Text = item.NoNota + " [VOID]"
					} else {
						text.Text = item.NoNota
					}
					text.Alignment = fyne.TextAlignCenter
				case 2:
					text.Text = item.Customer
					text.Alignment = fyne.TextAlignCenter
				case 3:
					text.Text = item.Total
					text.Alignment = fyne.TextAlignCenter
				case 4:
					text.Text = ""
					text.Hide()
					btn.Show()
					btnID := item.ID // capture variable safely
					btn.OnTapped = func() {
						sell, err := s.SellRepo.GetByID(btnID)
						if err != nil {
							dialog.ShowError(err, w)
							return
						}
						displayItems := LoadSellDisplayItems(s, sell.Details)
						err = PrintNotaPenjualan(sell.Header, displayItems)
						if err != nil {
							dialog.ShowError(fmt.Errorf("Gagal print nota: %v", err), w)
						} else {
							ShowSuccessToast("Success", "Nota berhasil dibuka!", w)
						}
					}
				}
			}
		},
	)

	table.SetColumnWidth(0, 130)
	table.SetColumnWidth(1, 230)
	table.SetColumnWidth(2, 310)
	table.SetColumnWidth(3, 170)
	table.SetColumnWidth(4, 90) // Aksi

	// focus helpers (re-used in many pages)
	var focusWrapper *focusableTable
	safeFocus := func() {
		if focusWrapper != nil {
			fyne.Do(func() {
				w.Canvas().Focus(focusWrapper)
			})
		}
	}

	// hook up search after table exists
	search.OnChanged = func(keyword string) {
		selectedRow = -1
		loadData(keyword)
		table.Refresh()
	}
	if focusWrapper != nil {
		fyne.Do(func() {
			w.Canvas().Focus(focusWrapper)
		})
	}

	// Refresh function
	refreshTable := func() {
		loadData(search.Text)
		table.Refresh()
		safeFocus()
	}

	var lastDialogTime time.Time

	// Keyboard shortcuts handler
	handleKey := func(k *fyne.KeyEvent) {
		if time.Since(lastDialogTime) < 500*time.Millisecond {
			return
		}

		switch k.Name {

		// New nota
		case fyne.KeyInsert:
			lastDialogTime = time.Now()
			showPenjualanDialog(w, s, refreshTable, nil)

		// Preview nota
		case fyne.KeyV:
			lastDialogTime = time.Now()
			if selectedRow >= 0 && selectedRow < len(data) {
				sell, err := s.SellRepo.GetByID(data[selectedRow].ID)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				showPenjualanDialog(w, s, refreshTable, sell)
			} else {
				dialog.ShowInformation("Info", "Pilih nota terlebih dahulu!", w)
			}
		case fyne.KeyDelete:
			lastDialogTime = time.Now()
			if selectedRow >= 0 && selectedRow < len(data) {
				selectedID := data[selectedRow].ID
				selectedStatus := data[selectedRow].Status

				if selectedStatus == "VOID" {
					dialog.ShowInformation("Info", "Nota ini sudah berstatus VOID!", w)
					return
				}

				dialog.ShowConfirm("Void Nota Penjualan",
					"Apakah Anda yakin ingin melakukan VOID pada nota ini?\n\nAksi ini akan memutarbalikan stok barang yang sudah dijual kembali ke sistem dan mengubah status nota menjadi VOID.",
					func(b bool) {
						if b {
							err := s.SellRepo.Void(selectedID, s.User.ID)
							if err != nil {
								dialog.ShowError(err, w)
							} else {
								dialog.ShowInformation("Sukses", "Nota berhasil di-Void!", w)
								refreshTable()
							}
						}
					}, w)
			} else {
				dialog.ShowInformation("Info", "Pilih nota terlebih dahulu sebelum di-Void!", w)
			}
		case fyne.KeyUp:
			if len(data) > 0 {
				if selectedRow > 0 {
					selectedRow--
				} else if selectedRow == -1 {
					selectedRow = 0
				}
				table.Refresh()
				table.ScrollTo(widget.TableCellID{Row: selectedRow + 1, Col: 0})
			}
		case fyne.KeyDown:
			if len(data) > 0 {
				if selectedRow < len(data)-1 {
					selectedRow++
				} else if selectedRow == -1 {
					selectedRow = 0
				}
				table.Refresh()
				table.ScrollTo(widget.TableCellID{Row: selectedRow + 1, Col: 0})
			}
		case fyne.KeyHome:
			if len(data) > 0 {
				selectedRow = 0
				table.Refresh()
				table.ScrollTo(widget.TableCellID{Row: 1, Col: 0})
			}
		case fyne.KeyEnd:
			if len(data) > 0 {
				selectedRow = len(data) - 1
				table.Refresh()
				table.ScrollTo(widget.TableCellID{Row: selectedRow + 1, Col: 0})
			}
		case fyne.KeyPageUp:
			if len(data) > 0 {
				if selectedRow == -1 {
					selectedRow = 0
				} else {
					selectedRow -= 10
					if selectedRow < 0 {
						selectedRow = 0
					}
				}
				table.Refresh()
				table.ScrollTo(widget.TableCellID{Row: selectedRow + 1, Col: 0})
			}
		case fyne.KeyPageDown:
			if len(data) > 0 {
				if selectedRow == -1 {
					selectedRow = 0
				} else {
					selectedRow += 10
					if selectedRow >= len(data) {
						selectedRow = len(data) - 1
					}
				}
				table.Refresh()
				table.ScrollTo(widget.TableCellID{Row: selectedRow + 1, Col: 0})
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
			fyne.NewSize(950, 480),
			focusWrapper,
		),
	)

	// Footer
	footer := canvas.NewText(
		"[Insert] Nota Baru  [V] Preview Nota  [Del] Void",
		color.White,
	)
	footer.TextStyle = fyne.TextStyle{Italic: true}

	// Content
	content := container.NewBorder(header, footer, nil, nil, tableWrapper)

	// Create a semi-transparent dark gray rectangle for the panel (matching login/home)
	rect := canvas.NewRectangle(color.NRGBA{R: 30, G: 30, B: 30, A: 180})
	rect.CornerRadius = 12
	rect.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 40} // Subtle white border
	rect.StrokeWidth = 1
	rect.SetMinSize(fyne.NewSize(1050, 650))

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
