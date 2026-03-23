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

type PembelianHeader struct {
	ID      uuid.UUID
	TglNota string
	NoNota  string
	Vendor  string
	Total   string
	Status  string
}

type PembelianItem struct {
	ItemID     uuid.UUID
	KodeBarang string
	NamaBarang string
	Qty        string
	Harga      string
	Total      string
}

func showPembelianDialog(w fyne.Window, s *state.Session, refreshCallback func(), existingData *models.PurchaseFull, isEditMode bool) {
	// Header form fields
	tglNota := widget.NewLabel(time.Now().Format("2006-01-02"))
	tglNota.TextStyle = fyne.TextStyle{Bold: true}

	// Calendar button with calendar icon only
	calendarBtn := widget.NewButtonWithIcon("", theme.CalendarIcon(), func() {
		ShowDatePickerDialog(w, tglNota.Text, func(selectedDate string) {
			tglNota.SetText(selectedDate)
		})
	})
	calendarBtn.Importance = widget.LowImportance

	tglNotaContainer := container.NewBorder(nil, nil, nil, calendarBtn, tglNota)
	
	// Use Labels in preview mode, Entries in edit/new mode
	var noNotaWidget fyne.CanvasObject
	var vendorWidget fyne.CanvasObject
	
	noNota := widget.NewEntry()
	vendor := widget.NewEntry()
	
	noNotaLabel := widget.NewLabel("")
	noNotaLabel.TextStyle = fyne.TextStyle{Bold: false}
	vendorLabel := widget.NewLabel("")
	vendorLabel.TextStyle = fyne.TextStyle{Bold: false}

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
	harga := widget.NewEntry()

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
	noNota.OnSubmitted = func(s string) { w.Canvas().Focus(vendor) }
	vendor.OnSubmitted = func(s string) { w.Canvas().Focus(kodeBarang) }
	kodeBarang.OnSubmitted = func(s string) { w.Canvas().Focus(qty) }
	qty.OnSubmitted = func(s string) { w.Canvas().Focus(harga) }

	var selectedItem *models.Item
	var selectedItemIndex int = -1
	var isSyncing bool

	// Map untuk menyimpan qty awal item di dalam nota (khusus mode edit)
	originalQtyMap := make(map[uuid.UUID]float64)
	if existingData != nil {
		for _, det := range existingData.Details {
			originalQtyMap[det.ItemID] += det.Qty
		}
	}

	clearItemForm := func() {
		kodeBarang.SetText("")
		namaBarang.SetText("")
		qty.SetText("")
		harga.SetText("")
		selectedItem = nil
		selectedItemIndex = -1
		stockInfo.Text = ""
		stockWarning.Text = ""
		stockInfo.Refresh()
		stockWarning.Refresh()
	}

	isSyncing = false

	namaBarang.OnChanged = func(value string) {
		if isSyncing {
			return
		}

		// always keep the select entry options in sync with our master list
		// accessing the internal field is not allowed, so we just reset
		// whenever the callback runs; SetOptions is cheap.
		namaBarang.SetOptions(itemOptions)

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
			stockInfo.Text = fmt.Sprintf("Stok tersedia: %.0f", item.Qty)
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
			v, err := strconv.ParseFloat(val, 64)
			// in pembelian, the stock goes UP. Wait, Pembelian = Purchasing from vendor.
			// Is stock validation needed for Pembelian? The user said:
			// "Tolong update warning yang kayak sebelumnya di retur pembelian buat di menu pembelian & penjualan juga dong"
			// Ok, I will just display the info and check if (maybe no check for pembelian since we are not taking it from inventory).
			// Let's just put info and warning logic. If the user wants a warning in Pembelian? Maybe not for "Stok kurang!" since we are adding stock.
			// The user just said "warning yang kayak sebelumnya".
			if err == nil {
				if v < 0 {
					stockWarning.Text = "Qty minus!"
				} else if isEditMode {
					originalQty := originalQtyMap[selectedItem.ID]
					if originalQty > 0 {
						reduction := originalQty - v
						if reduction > selectedItem.Qty {
							stockWarning.Text = "Stok tidak cukup (akan negatif)!"
						} else {
							stockWarning.Text = ""
						}
					} else {
						stockWarning.Text = ""
					}
				} else {
					stockWarning.Text = ""
				}
			} else {
				stockWarning.Text = ""
			}
			stockWarning.Refresh()
		}
	}

	var items []PembelianItem
	var itemsTable *widget.Table

	// =====================
	// Grand total UI
	// =====================
	totalLabel := canvas.NewText("Total : Rp 0", color.Black)
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}
	totalLabel.Alignment = fyne.TextAlignTrailing

	recalculateTotal := func() float64 {
		var sum float64
		for _, item := range items {
			val, err := ParseCurrencyString(item.Total)
			if err == nil {
				sum += val
			}
		}
		totalLabel.Text = "Total : " + FormatCurrency(sum)
		totalLabel.Refresh()
		return sum
	}

	// Header form definition - determine which widgets to use based on mode
	if existingData != nil && !isEditMode {
		// Preview mode: use Labels for disabled fields
		noNotaLabel.SetText(existingData.Header.PurchaseInvoiceNum)
		vendorLabel.SetText(existingData.Header.SupplierName)
		noNotaWidget = noNotaLabel
		vendorWidget = vendorLabel
		calendarBtn.Disable()
	} else {
		// New or Edit mode: use Entry fields
		noNotaWidget = noNota
		vendorWidget = vendor
	}
	
	headerForm := widget.NewForm(
		widget.NewFormItem("Tgl. Nota", tglNotaContainer),
		widget.NewFormItem("No. Nota", noNotaWidget),
		widget.NewFormItem("Vendor", vendorWidget),
	)
	headerFormSeparator := widget.NewSeparator()

	if existingData != nil {
		formattedDate := existingData.Header.PurchaseDate.Format("2006-01-02")
		tglNota.SetText(formattedDate)
		if isEditMode {
			noNota.SetText(existingData.Header.PurchaseInvoiceNum)
			vendor.SetText(existingData.Header.SupplierName)
		}

		displayItems := LoadPurchaseDisplayItems(s, existingData.Details)
		items = make([]PembelianItem, len(displayItems))
		for i, v := range displayItems {
			items[i] = PembelianItem{
				ItemID:     existingData.Details[i].ItemID,
				KodeBarang: v.Code,
				NamaBarang: v.Name,
				Qty:        v.Qty,
				Harga:      v.Price,
				Total:      v.Total,
			}
		}

		// Lock header inputs in edit mode only
		if isEditMode {
			// In edit mode, no additional locking needed for purchase date
		}
	}
	recalculateTotal()
	refreshItemsTable := func() {
		if itemsTable != nil {
			itemsTable.Refresh()
		}
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
		if kodeBarang.Text == "" {
			dialog.ShowInformation("Error", "Kode barang harus diisi!", w)
			return
		}
		if selectedItem == nil {
			// Try lookup one more time if not set
			item, err := s.ItemRepo.GetByCode(kodeBarang.Text)
			if err == nil {
				selectedItem = item
			} else {
				dialog.ShowInformation("Error", "Item tidak ditemukan!", w)
				return
			}
		}
		if qty.Text == "" || harga.Text == "" {
			dialog.ShowInformation("Error", "Qty dan harga harus diisi!", w)
			return
		}

		qtyVal, err := strconv.ParseFloat(qty.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Qty harus berupa angka!"), w)
			return
		}
		if qtyVal <= 0 {
			dialog.ShowInformation("Error", "Qty harus lebih besar dari 0!", w)
			return
		}

		hargaVal, err := ParseCurrencyString(harga.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Harga harus berupa angka!"), w)
			return
		}

		// Validasi stok negatif di pembelian (Edit Mode)
		if isEditMode {
			originalQty := originalQtyMap[selectedItem.ID]
			if originalQty > 0 {
				reduction := originalQty - qtyVal
				if reduction > selectedItem.Qty {
					dialog.ShowError(fmt.Errorf("Gagal Update: Pengurangan QTY (%v) melebihi stok yang ada (%v). Stok akan menjadi negatif!", reduction, selectedItem.Qty), w)
					return
				}
			}
		}

		total := qtyVal * hargaVal

		newItem := PembelianItem{
			ItemID:     selectedItem.ID,
			KodeBarang: kodeBarang.Text,
			NamaBarang: namaBarang.Text,
			Qty:        qty.Text,
			Harga:      FormatCurrency(hargaVal),
			Total:      FormatCurrency(total),
		}

		if selectedItemIndex >= 0 {
			items[selectedItemIndex] = newItem
		} else {
			items = append(items, newItem)
		}

		clearItemForm()
		updateButtonStates()
		refreshItemsTable()
		w.Canvas().Focus(kodeBarang)
	}

	harga.OnSubmitted = func(s string) { addItemBtn.OnTapped() }

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

	updateButtonStates()

	itemForm := widget.NewForm(
		widget.NewFormItem("Kode Barang", kodeBarang),
		widget.NewFormItem("Nama Barang", namaBarang),
		widget.NewFormItem("Qty", qtyContainer),
		widget.NewFormItem("Harga", harga),
	)
	itemFormSeparator := widget.NewSeparator()
	itemFormButtons := container.NewHBox(addItemBtn, deleteItemBtn, clearBtn)

	// Merged dialog: always show both forms
	// Removed initialFocus logic

	itemHeaders := []string{"Kode Barang", "Nama Barang", "QTY", "Harga", "Total", ""}
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	itemsTable = widget.NewTable(
		func() (int, int) { return len(items) + 1, len(itemHeaders) },
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
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
					if existingData != nil && !isEditMode {
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
		if existingData != nil && !isEditMode {
			return // Disable selection for existing data (read-only)
		}
		if id.Row > 0 && id.Row-1 < len(items) {
			selectedRow := id.Row - 1
			item := items[selectedRow]
			it, err := s.ItemRepo.GetByCode(item.KodeBarang)
			if err == nil {
				selectedItem = it
				selectedItemIndex = selectedRow
				isSyncing = true
				kodeBarang.SetText(item.KodeBarang)
				namaBarang.SetText(item.NamaBarang)
				isSyncing = false
				qty.SetText(item.Qty)
				harga.SetText(item.Harga)
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

	if existingData != nil && !isEditMode {
		kodeBarang.Disable()
		namaBarang.Disable()
		qty.Disable()
		harga.Disable()
		addItemBtn.Hide()
		deleteItemBtn.Hide()
		clearBtn.Hide()
		itemsTable.SetColumnWidth(4, 160) // Absorb Delete button width (120+40)
		itemsTable.SetColumnWidth(5, 0)
	}

	var d dialog.Dialog
	submitBtn := widget.NewButton("Submit", func() {
		if tglNota.Text == "" || noNota.Text == "" || vendor.Text == "" {
			dialog.ShowInformation("Error", "Header data harus diisi!", w)
			return
		}

		if !isEditMode {
			// Validasi No Nota duplikat
			existing, _ := s.PurchaseRepo.GetByInvoiceNum(noNota.Text)
			if existing != nil {
				dialog.ShowInformation("Error", "No. Nota sudah terdaftar, silakan gunakan nomor lain!", w)
				return
			}
		}

		// Validate item counts (prevent empty nota)
		if len(items) == 0 {
			dialog.ShowInformation("Error", "Minimal 1 item harus ditambahkan!", w)
			return
		}

		purchaseDate, _ := time.Parse("2006-01-02", tglNota.Text)
		grandTotal := recalculateTotal()

		purchase := &models.PurchaseFull{
			Header: models.PurchaseHeader{
				PurchaseInvoiceNum: noNota.Text,
				PurchaseDate:       purchaseDate,
				SupplierName:       vendor.Text,
				TotalAmount:        grandTotal,
			},
			Details: make([]models.PurchaseDetail, len(items)),
		}

		if existingData != nil {
			purchase.Header.ID = existingData.Header.ID
			purchase.Header.CreatedAt = existingData.Header.CreatedAt
			purchase.Header.CreatedBy = existingData.Header.CreatedBy
			now := time.Now()
			purchase.Header.UpdatedAt = &now
			purchase.Header.UpdatedBy = &s.User.ID
		} else {
			purchase.Header.CreatedBy = &s.User.ID
		}

		for i, item := range items {
			qtyVal, _ := strconv.ParseFloat(item.Qty, 64)
			hargaVal, _ := ParseCurrencyString(item.Harga)
			totalVal, _ := ParseCurrencyString(item.Total)
			purchase.Details[i] = models.PurchaseDetail{
				ItemID:      item.ItemID,
				Qty:         qtyVal,
				PriceAmount: hargaVal,
				TotalAmount: totalVal,
			}
		}

		var err error
		if isEditMode && existingData != nil {
			err = s.PurchaseRepo.Update(purchase)
		} else {
			err = s.PurchaseRepo.Create(purchase)
		}

		if err != nil {
			dialog.ShowError(fmt.Errorf("Gagal menyimpan data: %v", err), w)
			return
		}

		// mark blue if date changed on save
		// (UI color feedback removed)

		ShowSuccessToast("Success", "Data pembelian berhasil disimpan!", w)
		d.Hide()
		if refreshCallback != nil {
			refreshCallback()
		}
	})
	submitBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
		if refreshCallback != nil {
			refreshCallback()
		}
	})
	cancelBtn.Importance = widget.DangerImportance

	var buttons *fyne.Container

	if existingData != nil && !isEditMode {
		cancelBtn.SetText("Tutup")
		cancelBtn.Importance = widget.HighImportance
		buttons = container.NewGridWithColumns(1, cancelBtn)
		submitBtn.Hide()
	} else {
		buttons = container.NewGridWithColumns(2, cancelBtn, submitBtn)
	}

	labelText := "Nota Baru"
	if isEditMode {
		labelText = "Edit Nota"
	} else if existingData != nil {
		labelText = "Detail Nota"
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
			"MENU PEMBELIAN BARANG",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		)),
		container.NewCenter(widget.NewLabelWithStyle(
			labelText,
			fyne.TextAlignCenter,
			fyne.TextStyle{},
		)),
		widget.NewSeparator(),
	)

	if existingData != nil && !isEditMode {
		topContent.Add(headerForm)
		topContent.Add(headerFormSeparator)
		// Hide item input form entirely
	} else {
		topContent.Add(headerForm)
		topContent.Add(headerFormSeparator)
		topContent.Add(itemForm)
		topContent.Add(itemFormButtons)
		topContent.Add(itemFormSeparator)
	}

	content := container.NewBorder(
		// TOP
		topContent,

		// BOTTOM  👈 buttons must be here
		buttons,

		// LEFT / RIGHT
		nil,
		nil,

		// CENTER 👈 table + total must expand
		tableSection,
	)

	dialogContent := container.NewPadded(content)

	d = dialog.NewCustom("", "", dialogContent, w)
	d.Resize(fyne.NewSize(750, 600))
	d.Show()

	// Initial focus on Kode Barang first to encourage item adding
	if !(existingData != nil && !isEditMode) {
		time.AfterFunc(100*time.Millisecond, func() {
			fyne.Do(func() {
				w.Canvas().Focus(kodeBarang)
			})
		})
	}
}

// (Edit dialog removed to unify Add flow)

func showViewPembelianDialog(w fyne.Window, s *state.Session, headerID uuid.UUID) {
	purchase, err := s.PurchaseRepo.GetByID(headerID)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Gagal memuat data: %v", err), w)
		return
	}

	tglNota := widget.NewLabel(purchase.Header.PurchaseDate.Format("2006-01-02"))
	noNota := widget.NewLabel(purchase.Header.PurchaseInvoiceNum)
	vendor := widget.NewLabel(purchase.Header.SupplierName)

	headerInfo := container.NewVBox(
		container.NewGridWithColumns(2, canvas.NewText("Tgl. Nota", color.White), tglNota),
		container.NewGridWithColumns(2, canvas.NewText("No. Nota", color.White), noNota),
		container.NewGridWithColumns(2, canvas.NewText("Vendor", color.White), vendor),
	)

	displayItems := LoadPurchaseDisplayItems(s, purchase.Details)

	totalLabel := canvas.NewText(
		"Total : "+FormatCurrency(purchase.Header.TotalAmount),
		color.Black,
	)
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}
	totalLabel.Alignment = fyne.TextAlignTrailing

	itemHeaders := []string{"Kode Barang", "Nama Barang", "QTY", "Harga", "Total"}
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	itemsTable := widget.NewTable(
		func() (int, int) { return len(displayItems) + 1, len(itemHeaders) },
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
				text.Text = item.Price
				text.Alignment = fyne.TextAlignTrailing
			case 4:
				text.Text = item.Total
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
	closeBtn := widget.NewButton("Close", func() { d.Hide() })

	content := container.NewBorder(
		container.NewVBox(
			container.NewCenter(widget.NewLabelWithStyle("MENU PEMBELIAN BARANG", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
			container.NewCenter(widget.NewLabelWithStyle("View data", fyne.TextAlignCenter, fyne.TextStyle{})),
			widget.NewSeparator(),
			headerInfo,
			widget.NewSeparator(),
		),
		container.NewCenter(closeBtn), nil, nil, container.NewBorder(
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
		),
	)

	dialogContent := container.NewPadded(content)

	d = dialog.NewCustom("", "", dialogContent, w)
	d.Resize(fyne.NewSize(700, 500))
	d.Show()
}

func PembelianPage(w fyne.Window, s *state.Session) fyne.CanvasObject {
	bg := canvas.NewImageFromFile("assets/bg-login.jpg")
	bg.FillMode = canvas.ImageFillStretch

	backBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		w.SetContent(HomePage(w, s))
	})

	title := canvas.NewText("MENU PEMBELIAN BARANG", color.White)
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 16

	search := widget.NewEntry()
	search.SetPlaceHolder("Search No. Nota or Vendor...")

	// statusFilter := widget.NewSelect([]string{"Semua", "ACTIVE", "VOID"}, nil)
	// statusFilter.SetSelected("Semua")

	// rightPanel := container.NewGridWithColumns(2, search, statusFilter)
	// header := container.NewGridWithColumns(3, backBtn, title, container.NewMax(rightPanel))
	header := container.NewGridWithColumns(3, backBtn, title, container.NewMax(search))

	headers := []string{"Tgl. Nota", "No. Nota", "Vendor", "Total"}
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	var data []PembelianHeader
	var selectedRow int = -1
	var refreshTable func()
	var isDialogOpen bool

	// loadData := func(keyword string, sfilt string) {
	loadData := func(keyword string) {
		var headers []models.PurchaseHeader
		var err error
		if keyword == "" {
			headers, err = s.PurchaseRepo.GetAll()
		} else {
			headers, err = s.PurchaseRepo.Search(keyword)
		}
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		data = nil
		for _, h := range headers {
			// if sfilt != "Semua" && h.Status != sfilt {
			// 	continue
			// }
			data = append(data, PembelianHeader{
				ID:      h.ID,
				TglNota: h.PurchaseDate.Format("2006-01-02"),
				NoNota:  h.PurchaseInvoiceNum,
				Vendor:  h.SupplierName,
				Total:   FormatCurrency(h.TotalAmount),
				Status:  h.Status,
			})
		}
	}

	// loadData("", "Semua")
	loadData("")

	table := widget.NewTable(
		func() (int, int) { return len(data) + 1, len(headers) },
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			text := canvas.NewText("", color.Black)
			text.TextSize = 13
			
			editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			editBtn.Importance = widget.LowImportance
			return container.NewMax(bg, text, container.NewCenter(editBtn))
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			cont := cell.(*fyne.Container)
			bg := cont.Objects[0].(*canvas.Rectangle)
			text := cont.Objects[1].(*canvas.Text)
			
			btnCont := cont.Objects[2].(*fyne.Container)
			editBtn := btnCont.Objects[0].(*widget.Button)
			editBtn.Hide()
			text.Show()
			if id.Row == 0 {
				bg.FillColor = headerBg
				text.Text = headers[id.Col]
				text.Color = color.White
				text.Alignment = fyne.TextAlignCenter
				text.TextStyle = fyne.TextStyle{Bold: true}
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
					text.Text = item.Vendor
					text.Alignment = fyne.TextAlignCenter
				case 3:
					text.Text = item.Total
					text.Alignment = fyne.TextAlignCenter
				case 4:
					text.Text = ""
					text.Hide()
					editBtnID := item.ID
					editBtnStatus := item.Status
					if editBtnStatus == "VOID" {
						editBtn.Disable()
					} else {
						editBtn.Enable()
					}
					editBtn.OnTapped = func() {
						if editBtnStatus == "VOID" {
							dialog.ShowInformation("Info", "Nota VOID tidak bisa diedit!", w)
							return
						}
						purchase, err := s.PurchaseRepo.GetByID(editBtnID)
						if err != nil {
							dialog.ShowError(err, w)
							return
						}
						isDialogOpen = true
						showPembelianDialog(w, s, func() { isDialogOpen = false; refreshTable() }, purchase, true)
					}
					editBtn.Show()
				}
			}
		},
	)

	table.SetColumnWidth(0, 150)
	table.SetColumnWidth(1, 240)
	table.SetColumnWidth(2, 355)
	table.SetColumnWidth(3, 190)
	table.SetColumnWidth(4, 50) // Edit button

	// enable live searching
	search.OnChanged = func(keyword string) {
		selectedRow = -1
		loadData(keyword)
		table.Refresh()
	}
	// statusFilter.OnChanged = func(_ string) {
	// 	selectedRow = -1
	// 	loadData(search.Text, statusFilter.Selected)
	// 	table.Refresh()
	// }

	var focusWrapper *focusableTable
	safeFocus := func() {
		if focusWrapper != nil {
			fyne.Do(func() {
				w.Canvas().Focus(focusWrapper)
			})
		}
	}

	refreshTable = func() {
		loadData(search.Text)
		table.Refresh()
		safeFocus()
	}

	var lastDialogTime time.Time

	handleKey := func(k *fyne.KeyEvent) {
		if time.Since(lastDialogTime) < 500*time.Millisecond || isDialogOpen {
			return
		}

		switch k.Name {

		// New nota
		case fyne.KeyInsert:
			lastDialogTime = time.Now()
			isDialogOpen = true
			showPembelianDialog(w, s, func() { isDialogOpen = false; refreshTable() }, nil, false)

		// Edit nota
		case fyne.KeyE:
			lastDialogTime = time.Now()
			if selectedRow >= 0 && selectedRow < len(data) {
				if data[selectedRow].Status == "VOID" {
					dialog.ShowInformation("Info", "Nota VOID tidak bisa diedit!", w)
					return
				}
				purchase, err := s.PurchaseRepo.GetByID(data[selectedRow].ID)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				isDialogOpen = true
				showPembelianDialog(w, s, func() { isDialogOpen = false; refreshTable() }, purchase, true)
			} else {
				dialog.ShowInformation("Info", "Pilih nota terlebih dahulu!", w)
			}

		// Preview
		case fyne.KeyV:
			lastDialogTime = time.Now()
			if selectedRow >= 0 && selectedRow < len(data) {
				purchase, err := s.PurchaseRepo.GetByID(data[selectedRow].ID)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				isDialogOpen = true
				showPembelianDialog(w, s, func() { isDialogOpen = false; refreshTable() }, purchase, false)
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

				dialog.ShowConfirm("Void Nota Pembelian",
					"Apakah Anda yakin ingin melakukan VOID pada nota ini?\n\nAksi ini akan memutarbalikan stok barang yang sudah dibeli dan mengubah status nota menjadi VOID.",
					func(b bool) {
						if b {
							err := s.PurchaseRepo.Void(selectedID, s.User.ID)
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

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 {
			selectedRow = id.Row - 1
			table.Refresh()
			time.AfterFunc(50*time.Millisecond, safeFocus)
		}
	}

	focusWrapper = newFocusableTable(table, handleKey)
	w.Canvas().SetOnTypedKey(handleKey)

	tableWrapper := container.NewCenter(container.NewGridWrap(fyne.NewSize(950, 480), focusWrapper))
	footer := canvas.NewText("[Insert] Nota Baru  [V] Preview Nota  [Del] Void", color.White)
	footer.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewBorder(header, footer, nil, nil, tableWrapper)
	rect := canvas.NewRectangle(color.NRGBA{R: 30, G: 30, B: 30, A: 180})
	rect.CornerRadius = 12
	rect.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 40}
	rect.StrokeWidth = 1
	rect.SetMinSize(fyne.NewSize(1050, 650))

	panel := container.NewMax(rect, container.NewPadded(content))
	centeredPanel := container.NewCenter(panel)

	time.AfterFunc(150*time.Millisecond, safeFocus)

	return container.NewMax(bg, centeredPanel)
}
