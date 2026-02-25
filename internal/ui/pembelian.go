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

type PembelianHeader struct {
	ID      uuid.UUID
	TglNota string
	NoNota  string
	Vendor  string
}

type PembelianItem struct {
	ItemID     uuid.UUID
	KodeBarang string
	NamaBarang string
	Qty        string
	Harga      string
	Total      string
}

func showPembelianDialog(w fyne.Window, s *state.Session, refreshCallback func(), initialFocus string, existingData *models.PurchaseFull) {
	// Header form fields
	tglNota := widget.NewEntry()
	tglNota.SetText(time.Now().Format("2006-01-02"))

	calendarBtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		dateEntry := widget.NewEntry()
		dateEntry.SetText(tglNota.Text)
		dateDialog := dialog.NewCustomConfirm("Pilih Tanggal", "OK", "Cancel", container.NewVBox(widget.NewLabel("Format: YYYY-MM-DD"), dateEntry), func(ok bool) {
			if ok {
				tglNota.SetText(dateEntry.Text)
			}
		}, w)
		dateDialog.Show()
	})

	tglNotaContainer := container.NewBorder(nil, nil, nil, calendarBtn, tglNota)
	noNota := widget.NewEntry()
	vendor := widget.NewEntry()

	// Item entry fields
	kodeBarang := widget.NewEntry()
	namaBarang := widget.NewEntry()
	qty := widget.NewEntry()
	harga := widget.NewEntry()

	// Focus flow
	noNota.OnSubmitted = func(s string) { w.Canvas().Focus(vendor) }
	vendor.OnSubmitted = func(s string) { w.Canvas().Focus(kodeBarang) }
	kodeBarang.OnSubmitted = func(s string) { w.Canvas().Focus(qty) }
	qty.OnSubmitted = func(s string) { w.Canvas().Focus(harga) }

	var selectedItem *models.Item
	var selectedItemIndex int = -1

	clearItemForm := func() {
		kodeBarang.SetText("")
		namaBarang.SetText("")
		qty.SetText("")
		harga.SetText("")
		selectedItem = nil
		selectedItemIndex = -1
	}

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

	var items []PembelianItem
	var itemsTable *widget.Table

	if existingData != nil {
		tglNota.SetText(existingData.Header.PurchaseDate.Format("2006-01-02"))
		noNota.SetText(existingData.Header.PurchaseInvoiceNum)
		vendor.SetText(existingData.Header.SupplierName)

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
	}

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

		hargaVal, err := ParseCurrencyString(harga.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Harga harus berupa angka!"), w)
			return
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

	headerForm := widget.NewForm(
		widget.NewFormItem("Tgl. Nota", tglNotaContainer),
		widget.NewFormItem("No. Nota", noNota),
		widget.NewFormItem("Vendor", vendor),
	)
	headerFormSeparator := widget.NewSeparator()

	itemForm := widget.NewForm(
		widget.NewFormItem("Kode Barang", kodeBarang),
		widget.NewFormItem("Nama Barang", namaBarang),
		widget.NewFormItem("Qty", qty),
		widget.NewFormItem("Harga", harga),
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
		if initialFocus == "item" && len(items) == 0 {
			dialog.ShowInformation("Error", "Minimal 1 item harus ditambahkan!", w)
			return
		}

		purchaseDate, _ := time.Parse("2006-01-02", tglNota.Text)
		purchase := &models.PurchaseFull{
			Header: models.PurchaseHeader{
				PurchaseInvoiceNum: noNota.Text,
				PurchaseDate:       purchaseDate,
				SupplierName:       vendor.Text,
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
		if existingData != nil {
			err = s.PurchaseRepo.Update(purchase)
		} else {
			err = s.PurchaseRepo.Create(purchase)
		}

		if err != nil {
			dialog.ShowError(fmt.Errorf("Gagal menyimpan data: %v", err), w)
			return
		}

		ShowSuccessToast("Success", "Data pembelian berhasil disimpan!", w)
		d.Hide()
		if refreshCallback != nil {
			refreshCallback()
		}
	})
	submitBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() { d.Hide() })
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
			container.NewCenter(widget.NewLabelWithStyle("MENU PEMBELIAN BARANG", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
			container.NewCenter(widget.NewLabelWithStyle(labelText, fyne.TextAlignCenter, fyne.TextStyle{})),
			widget.NewSeparator(),
			headerForm,
			headerFormSeparator,
			itemForm,
			itemFormButtons,
			itemFormSeparator,
		),
		buttons, nil, nil, tableScroll,
	)

	dialogContent := container.NewPadded(content)

	d = dialog.NewCustom("", "", dialogContent, w)
	d.Resize(fyne.NewSize(650, 600))
	d.Show()

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

func showEditPembelianDialog(w fyne.Window, s *state.Session, headerID uuid.UUID, refreshCallback func()) {
	purchase, err := s.PurchaseRepo.GetByID(headerID)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	showPembelianDialog(w, s, refreshCallback, "item", purchase)
}

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
		container.NewCenter(closeBtn), nil, nil, container.NewScroll(itemsTable),
	)

	dialogContent := container.NewPadded(content)

	d = dialog.NewCustom("", "", dialogContent, w)
	d.Resize(fyne.NewSize(650, 500))
	d.Show()
}

func PembelianPage(w fyne.Window, s *state.Session) fyne.CanvasObject {
	bg := canvas.NewImageFromFile("assets/bg-login.jpg")
	bg.FillMode = canvas.ImageFillStretch

	backBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		w.SetContent(HomePage(w, s))
	})
	backBtn.Importance = widget.LowImportance

	title := canvas.NewText("MENU PEMBELIAN BARANG", color.White)
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 16

	search := widget.NewEntry()
	search.SetPlaceHolder("Search...")
	header := container.NewGridWithColumns(3, backBtn, title, container.NewMax(search))

	headers := []string{"Tgl. Nota", "No. Nota", "Vendor"}
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	var data []PembelianHeader
	var selectedRow int = -1

	loadData := func(keyword string) {
		selectedRow = -1
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
		data = make([]PembelianHeader, len(headers))
		for i, h := range headers {
			data[i] = PembelianHeader{
				ID:      h.ID,
				TglNota: h.PurchaseDate.Format("2006-01-02"),
				NoNota:  h.PurchaseInvoiceNum,
				Vendor:  h.SupplierName,
			}
		}
	}

	loadData("")

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
				switch id.Col {
				case 0:
					text.Text = item.TglNota
					text.Alignment = fyne.TextAlignCenter
				case 1:
					text.Text = item.NoNota
					text.Alignment = fyne.TextAlignCenter
				case 2:
					text.Text = item.Vendor
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

	refreshTable := func() {
		loadData(search.Text)
		table.Refresh()
		safeFocus()
	}

	handleKey := func(k *fyne.KeyEvent) {
		switch k.Name {
		case fyne.KeyInsert:
			showPembelianDialog(w, s, refreshTable, "header", nil)
		case fyne.KeyE:
			if selectedRow >= 0 && selectedRow < len(data) {
				showEditPembelianDialog(w, s, data[selectedRow].ID, refreshTable)
			} else {
				dialog.ShowInformation("Info", "Pilih nota terlebih dahulu!", w)
			}
		case fyne.KeyV:
			if selectedRow >= 0 && selectedRow < len(data) {
				showViewPembelianDialog(w, s, data[selectedRow].ID)
			} else {
				dialog.ShowInformation("Info", "Pilih data terlebih dahulu!", w)
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

	tableWrapper := container.NewCenter(container.NewGridWrap(fyne.NewSize(780, 400), focusWrapper))
	footer := canvas.NewText("[Insert] Nota Baru  [E] Isi Nota ", color.White)
	footer.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewBorder(header, footer, nil, nil, tableWrapper)
	rect := canvas.NewRectangle(color.NRGBA{R: 30, G: 30, B: 30, A: 180})
	rect.CornerRadius = 12
	rect.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 40}
	rect.StrokeWidth = 1
	rect.SetMinSize(fyne.NewSize(980, 560))

	panel := container.NewMax(rect, container.NewPadded(content))
	centeredPanel := container.NewCenter(panel)

	time.AfterFunc(150*time.Millisecond, safeFocus)

	return container.NewMax(bg, centeredPanel)
}
