package ui

import (
	"fmt"
	"image/color"
	"strconv"

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

type InventoryItem struct {
	ID        uuid.UUID
	Code      string
	Name      string
	Qty       string
	HargaDus  string
	HargaPack string
	HargaRent string
}

func showAddInventoryDialog(w fyne.Window, s *state.Session, refreshCallback func()) {

	kode := widget.NewEntry()
	nama := widget.NewEntry()
	qty := widget.NewEntry()
	hargaDus := widget.NewEntry()
	hargaPack := widget.NewEntry()
	hargaRent := widget.NewEntry()

	form := widget.NewForm(
		widget.NewFormItem("Kode", kode),
		widget.NewFormItem("Nama", nama),
		widget.NewFormItem("Qty", qty),
		widget.NewFormItem("Harga Dus", hargaDus),
		widget.NewFormItem("Harga Pack", hargaPack),
		widget.NewFormItem("Harga Rent", hargaRent),
	)

	// WHITE BACKGROUND FOR DIALOG
	bg := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	content := container.NewPadded(form)

	formContent := container.NewMax(
		bg,
		content,
	)

	// Create the dialog first (will set buttons later)
	var d dialog.Dialog

	// Create custom buttons with specific colors
	submitBtn := widget.NewButton("Submit", func() {
		// Validate input
		if kode.Text == "" || nama.Text == "" || qty.Text == "" || 
		   hargaDus.Text == "" || hargaPack.Text == "" || hargaRent.Text == "" {
			dialog.ShowError(fmt.Errorf("Semua field harus diisi!"), w)
			return
		}

		// Parse values
		qtyVal, err := strconv.ParseFloat(qty.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Qty harus berupa angka!"), w)
			return
		}

		hargaDusVal, err := strconv.ParseFloat(hargaDus.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Harga Dus harus berupa angka!"), w)
			return
		}

		hargaPackVal, err := strconv.ParseFloat(hargaPack.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Harga Pack harus berupa angka!"), w)
			return
		}

		hargaRentVal, err := strconv.ParseFloat(hargaRent.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Harga Rent harus berupa angka!"), w)
			return
		}

		// Create item model
		item := &models.Item{
			Code:      kode.Text,
			Name:      nama.Text,
			Qty:       qtyVal,
			BoxPrice:  hargaDusVal,
			PackPrice: hargaPackVal,
			RentPrice: hargaRentVal,
			CreatedBy: &s.User.ID,
		}

		// Save to database
		err = s.ItemRepo.Create(item)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Gagal menyimpan data: %v", err), w)
			return
		}

		dialog.ShowInformation("Success", "Data berhasil disimpan!", w)
		d.Hide()
		
		if refreshCallback != nil {
			refreshCallback()
		}
	})

	submitBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
	})
	
	// Make cancel button red
	cancelBtn.Importance = widget.DangerImportance

	// Create button container
	buttons := container.NewGridWithColumns(2, cancelBtn, submitBtn)
	
	// Combine form content with buttons
	finalContent := container.NewBorder(
		nil,      // top
		buttons,  // bottom
		nil,      // left
		nil,      // right
		formContent, // center
	)

	// Add background to entire dialog content
	dialogContent := container.NewMax(
		canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		container.NewPadded(finalContent),
	)

	// Create the dialog with custom content
	d = dialog.NewCustom(
		"Add new data",
		"",  // Empty dismiss label since we have custom buttons
		dialogContent,
		w,
	)

	d.Resize(fyne.NewSize(420, 400))
	d.Show()
}

func showEditInventoryDialog(w fyne.Window, s *state.Session, item InventoryItem, refreshCallback func()) {

	kode := widget.NewEntry()
	kode.SetText(item.Code)
	kode.Disable() // Code should not be editable
	
	nama := widget.NewEntry()
	nama.SetText(item.Name)
	
	qty := widget.NewEntry()
	qty.SetText(item.Qty)
	
	hargaDus := widget.NewEntry()
	hargaDus.SetText(item.HargaDus)
	
	hargaPack := widget.NewEntry()
	hargaPack.SetText(item.HargaPack)
	
	hargaRent := widget.NewEntry()
	hargaRent.SetText(item.HargaRent)

	form := widget.NewForm(
		widget.NewFormItem("Kode", kode),
		widget.NewFormItem("Nama", nama),
		widget.NewFormItem("Qty", qty),
		widget.NewFormItem("Harga Dus", hargaDus),
		widget.NewFormItem("Harga Pack", hargaPack),
		widget.NewFormItem("Harga Rent", hargaRent),
	)

	// WHITE BACKGROUND FOR DIALOG
	bg := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	content := container.NewPadded(form)

	formContent := container.NewMax(
		bg,
		content,
	)

	// Create the dialog first (will set buttons later)
	var d dialog.Dialog

	// Create custom buttons with specific colors
	submitBtn := widget.NewButton("Submit", func() {
		// Validate input
		if nama.Text == "" || qty.Text == "" || 
		   hargaDus.Text == "" || hargaPack.Text == "" || hargaRent.Text == "" {
			dialog.ShowError(fmt.Errorf("Semua field harus diisi!"), w)
			return
		}

		// Parse values
		qtyVal, err := strconv.ParseFloat(qty.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Qty harus berupa angka!"), w)
			return
		}

		hargaDusVal, err := strconv.ParseFloat(hargaDus.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Harga Dus harus berupa angka!"), w)
			return
		}

		hargaPackVal, err := strconv.ParseFloat(hargaPack.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Harga Pack harus berupa angka!"), w)
			return
		}

		hargaRentVal, err := strconv.ParseFloat(hargaRent.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Harga Rent harus berupa angka!"), w)
			return
		}

		// Update item model
		updatedItem := &models.Item{
			ID:        item.ID,
			Code:      item.Code,
			Name:      nama.Text,
			Qty:       qtyVal,
			BoxPrice:  hargaDusVal,
			PackPrice: hargaPackVal,
			RentPrice: hargaRentVal,
			UpdatedBy: &s.User.ID,
		}

		// Update in database
		err = s.ItemRepo.Update(updatedItem)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Gagal mengupdate data: %v", err), w)
			return
		}

		dialog.ShowInformation("Success", "Data berhasil diupdate!", w)
		d.Hide()
		
		if refreshCallback != nil {
			refreshCallback()
		}
	})

	submitBtn.Importance = widget.HighImportance
	
	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
	})
	
	// Make cancel button red
	cancelBtn.Importance = widget.DangerImportance

	// Create button container
	buttons := container.NewGridWithColumns(2, cancelBtn, submitBtn)
	
	// Combine form content with buttons
	finalContent := container.NewBorder(
		nil,      // top
		buttons,  // bottom
		nil,      // left
		nil,      // right
		formContent, // center
	)

	// Add background to entire dialog content
	dialogContent := container.NewMax(
		canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		container.NewPadded(finalContent),
	)

	// Create the dialog with custom content
	d = dialog.NewCustom(
		"Edit data",
		"",  // Empty dismiss label since we have custom buttons
		dialogContent,
		w,
	)

	d.Resize(fyne.NewSize(420, 400))
	d.Show()
}

func InventoryPage(w fyne.Window, s *state.Session) fyne.CanvasObject {

	// ===== BACKGROUND =====
	bg := canvas.NewImageFromFile("assets/bg-login.jpg")
	bg.FillMode = canvas.ImageFillStretch

	// ===== HEADER ELEMENTS =====
	backBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		w.SetContent(HomePage(w, s))
	})
	backBtn.Importance = widget.LowImportance

	title := widget.NewLabelWithStyle(
		"MENU INVENTORY / OPNAME",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	search := widget.NewEntry()
	search.SetPlaceHolder("Search barang...")

	header := container.NewGridWithColumns(
		3,
		backBtn,
		title,
		container.NewMax(search),
	)

	// ===== DATA =====
	headers := []string{
		"Kode Barang",
		"Nama Barang",
		"QTY",
		"Harga Dus",
		"Harga Pack",
		"Harga Rent",
	}

	var data []InventoryItem
	var selectedRow int = -1

	// Load data from database
	loadData := func(keyword string) {
		var items []models.Item
		var err error
		
		if keyword == "" {
			items, err = s.ItemRepo.GetAll()
		} else {
			items, err = s.ItemRepo.Search(keyword)
		}
		
		if err != nil {
			dialog.ShowError(fmt.Errorf("Gagal memuat data: %v", err), w)
			return
		}

		data = make([]InventoryItem, len(items))
		for i, item := range items {
			data[i] = InventoryItem{
				ID:        item.ID,
				Code:      item.Code,
				Name:      item.Name,
				Qty:       fmt.Sprintf("%.0f", item.Qty),
				HargaDus:  fmt.Sprintf("%.0f", item.BoxPrice),
				HargaPack: fmt.Sprintf("%.0f", item.PackPrice),
				HargaRent: fmt.Sprintf("%.0f", item.RentPrice),
			}
		}
	}

	// Initial load
	loadData("")

	// ===== COLORS =====
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	// ===== TABLE =====
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
			objects := cell.(*fyne.Container).Objects
			bg := objects[0].(*canvas.Rectangle)
			text := objects[1].(*canvas.Text)

			if id.Row == 0 {
				bg.FillColor = headerBg
				text.Text = headers[id.Col]
				text.Color = color.White
				text.TextSize = 14
				text.TextStyle = fyne.TextStyle{Bold: true}
				text.Alignment = fyne.TextAlignCenter
				text.Refresh()
				return
			}

			bg.FillColor = rowBg
			text.Color = color.Black
			text.TextStyle = fyne.TextStyle{}
			text.TextSize = 13

			if id.Row-1 < len(data) {
				item := data[id.Row-1]
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
					text.Text = item.HargaDus
					text.Alignment = fyne.TextAlignTrailing
				case 4:
					text.Text = item.HargaPack
					text.Alignment = fyne.TextAlignTrailing
				case 5:
					text.Text = item.HargaRent
					text.Alignment = fyne.TextAlignTrailing
				}
			}

			text.Refresh()
		},
	)

	// Table selection
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 {
			selectedRow = id.Row - 1
		}
	}

	// ===== COLUMN WIDTH (FIXED & BALANCED) =====
	table.SetColumnWidth(0, 110)
	table.SetColumnWidth(1, 200)
	table.SetColumnWidth(2, 70)
	table.SetColumnWidth(3, 126)
	table.SetColumnWidth(4, 126)
	table.SetColumnWidth(5, 126)

	// Search functionality
	search.OnChanged = func(keyword string) {
		loadData(keyword)
		table.Refresh()
	}

	// Refresh callback
	refreshTable := func() {
		loadData(search.Text)
		table.Refresh()
	}

	// ===== FOOTER =====
	footer := widget.NewLabelWithStyle(
		"Insert = input data   |   E = Edit data   |   Del = Delete data",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	// ===== CENTERED TABLE WRAPPER =====
	tableWrapper := container.NewCenter(
		container.NewGridWrap(
			fyne.NewSize(780, 400),
			table,
		),
	)

	// ===== CONTENT =====
	content := container.NewBorder(
		header,
		footer,
		nil,
		nil,
		tableWrapper,
	)

	card := widget.NewCard("", "", content)

	// Keyboard shortcuts
	w.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		switch k.Name {
		case fyne.KeyInsert:
			showAddInventoryDialog(w, s, refreshTable)
		case fyne.KeyE:
			if selectedRow >= 0 && selectedRow < len(data) {
				showEditInventoryDialog(w, s, data[selectedRow], refreshTable)
			} else {
				dialog.ShowInformation("Info", "Pilih data terlebih dahulu!", w)
			}
		case fyne.KeyDelete:
			if selectedRow >= 0 && selectedRow < len(data) {
				selectedItem := data[selectedRow]
				dialog.ShowConfirm("Konfirmasi", 
					fmt.Sprintf("Apakah Anda yakin ingin menghapus '%s'?", selectedItem.Name),
					func(confirmed bool) {
						if confirmed {
							err := s.ItemRepo.Delete(selectedItem.ID, s.User.ID)
							if err != nil {
								dialog.ShowError(fmt.Errorf("Gagal menghapus data: %v", err), w)
								return
							}
							dialog.ShowInformation("Success", "Data berhasil dihapus!", w)
							refreshTable()
						}
					}, w)
			} else {
				dialog.ShowInformation("Info", "Pilih data terlebih dahulu!", w)
			}
		}
	})

	return container.NewMax(
		bg,
		container.NewCenter(
			container.NewGridWrap(
				fyne.NewSize(980, 560),
				card,
			),
		),
	)
}