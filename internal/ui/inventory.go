package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/dialog"
	"fyne-app/internal/state"
)

type InventoryItem struct {
	Code      string
	Name      string
	Qty       string
	HargaDus  string
	HargaPack string
	HargaRent string
}

func showAddInventoryDialog(w fyne.Window) {

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
		// TODO:
		// - validate input
		// - append to data
		// - refresh table
		d.Hide()
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

func showEditInventoryDialog(w fyne.Window) {

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
		// TODO:
		// - validate input
		// - append to data
		// - refresh table
		d.Hide()
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

	data := []InventoryItem{
		{"BRG001", "Indomie Goreng", "120", "120.000", "10.000", "1.200"},
		{"BRG002", "Aqua 600ml", "80", "90.000", "7.500", "900"},
		{"BRG003", "Kopi Sachet", "200", "60.000", "5.000", "600"},
	}

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

			text.Refresh()
		},
	)

	// ===== COLUMN WIDTH (FIXED & BALANCED) =====
	table.SetColumnWidth(0, 110)
	table.SetColumnWidth(1, 200)
	table.SetColumnWidth(2, 70)
	table.SetColumnWidth(3, 126)
	table.SetColumnWidth(4, 126)
	table.SetColumnWidth(5, 126)

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

	// Single call handles both keys
	w.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		switch k.Name {
		case fyne.KeyInsert:
			showAddInventoryDialog(w)
		case fyne.KeyE:
			showEditInventoryDialog(w)
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