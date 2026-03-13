package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"fyne-app/internal/models"
	"fyne-app/internal/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
)

type ReturHeaderUI struct {
	ID      uuid.UUID
	TglNota string
	NoNota  string
	Vendor  string
	Total   string
	Status  string
}

type ReturItemUI struct {
	ItemID     uuid.UUID
	KodeBarang string
	NamaBarang string
	Qty        string
	Harga      string
	Total      string
}

func showReturDialog(w fyne.Window, s *state.Session, refreshCallback func()) {
	// Header form fields
	tglNota := widget.NewEntry()
	tglNota.SetText(time.Now().Format("2006-01-02"))
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
	vendor := widget.NewEntry()

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
	noNota.OnSubmitted = func(st string) { w.Canvas().Focus(vendor) }
	vendor.OnSubmitted = func(st string) { w.Canvas().Focus(kodeBarang) }
	kodeBarang.OnSubmitted = func(st string) { w.Canvas().Focus(qty) }
	qty.OnSubmitted = func(st string) { w.Canvas().Focus(harga) }

	var selectedItem *models.Item
	var selectedItemIndex int = -1

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

	isSyncing := false

	namaBarang.OnChanged = func(value string) {
		if isSyncing {
			return
		}

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

	var items []ReturItemUI
	var itemsTable *widget.Table

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

			stockInfo.Text = fmt.Sprintf("Stok tersedia: %.0f", availableStock)
			stockInfo.Refresh()

			v, err := strconv.ParseFloat(val, 64)
			if err == nil {
				// Hitung berapa item ini yang sudah ada di cart
				if v > availableStock {
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

	// Items table variables moved above qty.OnChanged

	// =====================
	// Grand total UI
	// =====================
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
			val, err := ParseCurrencyString(item.Total)
			if err == nil {
				sum += val
			}
		}
		totalLabel.Text = "Total : " + FormatCurrency(sum)
		totalLabel.Refresh()
		return sum
	}

	refreshItemsTable := func() {
		if itemsTable != nil {
			itemsTable.Refresh()
		}
		recalculateTotal()
		updateStockInfo()
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
		if kodeBarang.Text == "" {
			dialog.ShowInformation("Error", "Kode barang harus diisi!", w)
			return
		}
		if selectedItem == nil {
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

		// Cek batas ketersediaan stok fisik gudang yang akan diretur (Retur = kembalikan stok/kurangi)
		var currentInCart float64
		for i, item := range items {
			if item.ItemID == selectedItem.ID && i != selectedItemIndex {
				q, _ := strconv.ParseFloat(item.Qty, 64)
				currentInCart += q
			}
		}

		availableStock := selectedItem.Qty - currentInCart

		if qtyVal > availableStock {
			dialog.ShowError(fmt.Errorf("Stok gudang tidak mencukupi untuk diretur! Sisa stok fisik: %.0f", availableStock), w)
			return
		}

		total := qtyVal * hargaVal

		newItem := ReturItemUI{
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

	harga.OnSubmitted = func(st string) { addItemBtn.OnTapped() }

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

	headerForm := widget.NewForm(
		widget.NewFormItem("Tgl. Nota", tglNotaContainer),
		widget.NewFormItem("No. Nota", noNota),
		widget.NewFormItem("Vendor", vendor),
	)
	headerFormSeparator := widget.NewSeparator()

	itemForm := widget.NewForm(
		widget.NewFormItem("Kode Barang", kodeBarang),
		widget.NewFormItem("Nama Barang", namaBarang),
		widget.NewFormItem("Qty", qtyContainer),
		widget.NewFormItem("Harga", harga),
	)
	itemFormSeparator := widget.NewSeparator()
	itemFormButtons := container.NewHBox(addItemBtn, deleteItemBtn, clearBtn)

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
			selectedRow := id.Row - 1
			item := items[selectedRow]
			it, err := s.ItemRepo.GetByCode(item.KodeBarang)
			if err == nil {
				selectedItem = it
				selectedItemIndex = selectedRow
				kodeBarang.SetText(item.KodeBarang)
				namaBarang.SetText(item.NamaBarang)
				qty.SetText(item.Qty)
				harga.SetText(item.Harga)

				stockInfo.Text = fmt.Sprintf("Stok tersedia: %.0f", it.Qty)
				stockInfo.Refresh()

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

	var d dialog.Dialog
	submitBtn := widget.NewButton("Submit", func() {
		if tglNota.Text == "" || noNota.Text == "" || vendor.Text == "" {
			dialog.ShowInformation("Error", "Header data harus diisi!", w)
			return
		}

		// Validasi No Nota duplikat
		existing, _ := s.ReturRepo.GetByInvoiceNum(noNota.Text)
		if existing != nil {
			dialog.ShowInformation("Error", "No. Nota sudah terdaftar, silakan gunakan nomor lain!", w)
			return
		}

		if len(items) == 0 {
			dialog.ShowInformation("Error", "Minimal 1 item harus ditambahkan!", w)
			return
		}

		returDate, _ := time.Parse("2006-01-02", tglNota.Text)
		grandTotal := recalculateTotal()

		header := models.ReturHeader{
			ReturInvoiceNum: noNota.Text,
			ReturDate:       returDate,
			SupplierName:    vendor.Text,
			TotalAmount:     grandTotal,
			CreatedBy:       &s.User.ID,
		}

		details := make([]models.ReturDetail, len(items))

		for i, item := range items {
			qtyVal, _ := strconv.ParseFloat(item.Qty, 64)
			hargaVal, _ := ParseCurrencyString(item.Harga)
			totalVal, _ := ParseCurrencyString(item.Total)
			details[i] = models.ReturDetail{
				ItemID:      item.ItemID,
				Qty:         qtyVal,
				PriceAmount: hargaVal,
				TotalAmount: totalVal,
			}
		}

		_, err := s.ReturRepo.Create(&header, details)

		if err != nil {
			dialog.ShowError(fmt.Errorf("Gagal menyimpan data: %v", err), w)
			return
		}

		ShowSuccessToast("Success", "Data retur pembelian berhasil disimpan!", w)
		d.Hide()
		if refreshCallback != nil {
			refreshCallback()
		}
	})
	submitBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() { d.Hide() })
	cancelBtn.Importance = widget.DangerImportance
	buttons := container.NewGridWithColumns(2, cancelBtn, submitBtn)

	labelText := "Isi Data Retur Pembelian"

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

	content := container.NewBorder(
		container.NewVBox(
			container.NewCenter(widget.NewLabelWithStyle(
				"MENU RETUR PEMBELIAN",
				fyne.TextAlignCenter,
				fyne.TextStyle{Bold: true},
			)),
			container.NewCenter(widget.NewLabelWithStyle(
				labelText,
				fyne.TextAlignCenter,
				fyne.TextStyle{},
			)),
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
		tableSection,
	)

	dialogContent := container.NewPadded(content)

	d = dialog.NewCustom("", "", dialogContent, w)
	d.Resize(fyne.NewSize(750, 600))
	d.Show()

	time.AfterFunc(100*time.Millisecond, func() {
		fyne.Do(func() {
			w.Canvas().Focus(noNota)
		})
	})
}

func showViewReturDialog(w fyne.Window, s *state.Session, headerID uuid.UUID) {
	retur, err := s.ReturRepo.GetByID(headerID)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Gagal memuat data: %v", err), w)
		return
	}

	tglNota := widget.NewLabel(retur.Header.ReturDate.Format("2006-01-02"))
	noNota := widget.NewLabel(retur.Header.ReturInvoiceNum)
	vendor := widget.NewLabel(retur.Header.SupplierName)

	headerInfo := widget.NewForm(
		widget.NewFormItem("Tgl. Nota", tglNota),
		widget.NewFormItem("No. Nota", noNota),
		widget.NewFormItem("Vendor", vendor),
	)

	displayItems := LoadReturDisplayItems(s, retur.Details)

	totalLabel := canvas.NewText(
		"Total : "+FormatCurrency(retur.Header.TotalAmount),
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

	itemsTable.SetColumnWidth(0, 110) // Kode
	itemsTable.SetColumnWidth(1, 170) // Nama
	itemsTable.SetColumnWidth(2, 70)  // Qty
	itemsTable.SetColumnWidth(3, 110) // Harga
	itemsTable.SetColumnWidth(4, 150) // Total (Absorb the 40px)

	var d dialog.Dialog
	closeBtn := widget.NewButton("Close", func() { d.Hide() })

	content := container.NewBorder(
		container.NewVBox(
			container.NewCenter(widget.NewLabelWithStyle("MENU RETUR PEMBELIAN", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
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

func ReturPage(w fyne.Window, s *state.Session) fyne.CanvasObject {
	bg := canvas.NewImageFromFile("assets/bg-login.jpg")
	bg.FillMode = canvas.ImageFillStretch

	backBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		w.SetContent(HomePage(w, s))
	})

	title := canvas.NewText("MENU RETUR PEMBELIAN", color.White)
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 16

	search := widget.NewEntry()
	search.SetPlaceHolder("Search No. Nota or Vendor...")

	statusFilter := widget.NewSelect([]string{"Semua", "ACTIVE", "VOID"}, nil)
	statusFilter.SetSelected("Semua")

	rightPanel := container.NewGridWithColumns(2, search, statusFilter)
	header := container.NewGridWithColumns(3, backBtn, title, container.NewMax(rightPanel))

	headers := []string{"Tgl. Nota", "No. Nota", "Vendor", "Total"}
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	var data []ReturHeaderUI
	var selectedRow int = -1

	loadData := func(keyword string, sfilt string) {
		selectedRow = -1
		var headers []models.ReturHeader
		var err error
		if keyword == "" {
			headers, err = s.ReturRepo.GetAll()
		} else {
			headers, err = s.ReturRepo.Search(keyword)
		}
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		data = nil
		for _, h := range headers {
			if sfilt != "Semua" && h.Status != sfilt {
				continue
			}
			data = append(data, ReturHeaderUI{
				ID:      h.ID,
				TglNota: h.ReturDate.Format("2006-01-02"),
				NoNota:  h.ReturInvoiceNum,
				Vendor:  h.SupplierName,
				Total:   FormatCurrency(h.TotalAmount),
				Status:  h.Status,
			})
		}
	}

	loadData("", "Semua")

	table := widget.NewTable(
		func() (int, int) { return len(data) + 1, len(headers) },
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
				}
			}
		},
	)

	table.SetColumnWidth(0, 150)
	table.SetColumnWidth(1, 240)
	table.SetColumnWidth(2, 290)
	table.SetColumnWidth(3, 240)

	search.OnChanged = func(keyword string) {
		selectedRow = -1
		loadData(keyword, statusFilter.Selected)
		table.Refresh()
	}
	statusFilter.OnChanged = func(_ string) {
		selectedRow = -1
		loadData(search.Text, statusFilter.Selected)
		table.Refresh()
	}

	var focusWrapper *focusableTable
	safeFocus := func() {
		if focusWrapper != nil {
			fyne.Do(func() {
				w.Canvas().Focus(focusWrapper)
			})
		}
	}

	refreshTable := func() {
		loadData(search.Text, statusFilter.Selected)
		table.Refresh()
		safeFocus()
	}

	var lastDialogTime time.Time

	handleKey := func(k *fyne.KeyEvent) {
		if time.Since(lastDialogTime) < 500*time.Millisecond {
			return
		}

		switch k.Name {

		case fyne.KeyInsert:
			lastDialogTime = time.Now()
			showReturDialog(w, s, refreshTable)

		// Show View details
		case fyne.KeyV:
			lastDialogTime = time.Now()
			if selectedRow >= 0 && selectedRow < len(data) {
				showViewReturDialog(w, s, data[selectedRow].ID)
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

				dialog.ShowConfirm("Void Nota Retur Pembelian",
					"Apakah Anda yakin ingin melakukan VOID pada nota retur ini?\n\nAksi ini akan membatalkan retur barang ke supplier sehingga barang akan masuk kembali ke stok gudang, dan mengubah status nota menjadi VOID.",
					func(b bool) {
						if b {
							err := s.ReturRepo.Void(selectedID, s.User.ID)
							if err != nil {
								dialog.ShowError(err, w)
							} else {
								dialog.ShowInformation("Sukses", "Nota Retur berhasil di-Void!", w)
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
	footer := canvas.NewText("[Insert] Add Retur  [V] View Detail  [Del] Void", color.White)
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
