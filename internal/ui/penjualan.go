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
	"fyne-app/internal/state"
)

type PenjualanHeader struct {
	TglNota  string
	NoNota   string
	Customer string
}

type PenjualanItem struct {
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

// Global data storage (replace with database later)
var penjualanData []PenjualanFull

func showAddPenjualanDialog(w fyne.Window, refreshCallback func()) {
	// Header form fields
	tglNota := widget.NewEntry()
	tglNota.SetText(time.Now().Format("2006-01-02"))
	
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
	
	// Harga dropdown (combo select with arrow)
	hargaOptions := []string{"Harga Dus", "Harga Pack", "Harga Rent"}
	harga := widget.NewSelect(hargaOptions, func(value string) {
		// Handle selection
	})
	harga.PlaceHolder = "Pilih Harga"

	// Items table data
	var items []PenjualanItem
	var itemsTable *widget.Table

	// Function to refresh items table
	refreshItemsTable := func() {
		if itemsTable != nil {
			itemsTable.Refresh()
		}
	}

	// Add item button
	addItemBtn := widget.NewButton("Add", func() {
		// Validate inputs
		if kodeBarang.Text == "" || namaBarang.Text == "" || qty.Text == "" || harga.Selected == "" {
			dialog.ShowInformation("Error", "Semua field item harus diisi!", w)
			return
		}

		// For demo, use dummy price values
		var hargaVal float64
		switch harga.Selected {
		case "Harga Dus":
			hargaVal = 100000 // Replace with actual lookup from inventory
		case "Harga Pack":
			hargaVal = 10000
		case "Harga Rent":
			hargaVal = 1000
		}

		// Calculate total
		qtyVal, _ := strconv.ParseFloat(qty.Text, 64)
		total := qtyVal * hargaVal

		// Add to items list
		items = append(items, PenjualanItem{
			KodeBarang: kodeBarang.Text,
			NamaBarang: namaBarang.Text,
			Qty:        qty.Text,
			Harga:      fmt.Sprintf("%.0f", hargaVal),
			Total:      fmt.Sprintf("%.0f", total),
		})

		// Clear entry fields
		kodeBarang.SetText("")
		namaBarang.SetText("")
		qty.SetText("")
		harga.Selected = ""
		harga.Refresh()

		// Refresh table
		refreshItemsTable()
	})

	// Header form
	headerForm := widget.NewForm(
		widget.NewFormItem("Tgl. Nota", tglNotaContainer),
		widget.NewFormItem("No. Nota", noNota),
		widget.NewFormItem("Customer", customer),
	)

	// Item entry form
	itemForm := container.NewGridWithColumns(2,
		widget.NewLabel("Kode Barang"),
		kodeBarang,
		widget.NewLabel("Nama Barang"),
		namaBarang,
		widget.NewLabel("Qty"),
		qty,
		widget.NewLabel("Harga"),
		harga,
	)

	// Items table
	itemHeaders := []string{"Kode Barang", "Nama Barang", "QTY", "Harga", "Total", ""}
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	itemsTable = widget.NewTable(
		func() (int, int) {
			return len(items) + 1, len(itemHeaders)
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			text := canvas.NewText("", color.Black)
			text.TextSize = 12
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
				// Header row
				bg.FillColor = headerBg
				text.Text = itemHeaders[id.Col]
				text.Color = color.White
				text.TextSize = 13
				text.TextStyle = fyne.TextStyle{Bold: true}
				text.Alignment = fyne.TextAlignCenter
				text.Show()
				text.Refresh()
				return
			}

			// Data rows
			bg.FillColor = rowBg
			text.Color = color.Black
			text.TextStyle = fyne.TextStyle{}
			text.TextSize = 12

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
					// Delete button
					text.Hide()
					btn.OnTapped = func() {
						rowIndex := id.Row - 1
						// Remove item
						items = append(items[:rowIndex], items[rowIndex+1:]...)
						refreshItemsTable()
					}
					btn.Show()
				}
			}
			text.Refresh()
		},
	)

	itemsTable.SetColumnWidth(0, 100)
	itemsTable.SetColumnWidth(1, 150)
	itemsTable.SetColumnWidth(2, 60)
	itemsTable.SetColumnWidth(3, 100)
	itemsTable.SetColumnWidth(4, 100)
	itemsTable.SetColumnWidth(5, 50)

	// Dialog content
	var d dialog.Dialog

	submitBtn := widget.NewButton("Submit", func() {
		// Validate
		if tglNota.Text == "" || noNota.Text == "" || customer.Text == "" {
			dialog.ShowInformation("Error", "Header data harus diisi!", w)
			return
		}
		if len(items) == 0 {
			dialog.ShowInformation("Error", "Minimal 1 item harus ditambahkan!", w)
			return
		}

		// Save data
		penjualanData = append(penjualanData, PenjualanFull{
			Header: PenjualanHeader{
				TglNota:  tglNota.Text,
				NoNota:   noNota.Text,
				Customer: customer.Text,
			},
			Items: items,
		})

		dialog.ShowInformation("Success", "Data penjualan berhasil disimpan!", w)
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

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("MENU PENJUALAN BARANG", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Add new data", fyne.TextAlignCenter, fyne.TextStyle{}),
			widget.NewSeparator(),
			headerForm,
			widget.NewSeparator(),
			itemForm,
			container.NewHBox(addItemBtn),
			widget.NewSeparator(),
		),
		buttons,
		nil,
		nil,
		container.NewScroll(itemsTable),
	)

	bg := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	dialogContent := container.NewMax(bg, container.NewPadded(content))

	d = dialog.NewCustom("", "", dialogContent, w)
	d.Resize(fyne.NewSize(650, 600))
	d.Show()
}

func showViewPenjualanDialog(w fyne.Window, penjualan PenjualanFull) {
	// Read-only header info
	tglNota := widget.NewLabel(penjualan.Header.TglNota)
	noNota := widget.NewLabel(penjualan.Header.NoNota)
	customer := widget.NewLabel(penjualan.Header.Customer)

	headerInfo := container.NewGridWithColumns(2,
		widget.NewLabel("Tgl. Nota"),
		tglNota,
		widget.NewLabel("No. Nota"),
		noNota,
		widget.NewLabel("Customer"),
		customer,
	)

	// Items table
	itemHeaders := []string{"Kode Barang", "Nama Barang", "QTY", "Harga", "Total"}
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	itemsTable := widget.NewTable(
		func() (int, int) {
			return len(penjualan.Items) + 1, len(itemHeaders)
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

			item := penjualan.Items[id.Row-1]
			switch id.Col {
			case 0:
				text.Text = item.KodeBarang
				text.Alignment = fyne.TextAlignLeading
			case 1:
				text.Text = item.NamaBarang
				text.Alignment = fyne.TextAlignLeading
			case 2:
				text.Text = item.Qty
				text.Alignment = fyne.TextAlignCenter
			case 3:
				text.Text = item.Harga
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

	closeBtn := widget.NewButton("Close", func() {
		d.Hide()
	})

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("MENU PENJUALAN BARANG", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("View data", fyne.TextAlignCenter, fyne.TextStyle{}),
			widget.NewSeparator(),
			headerInfo,
			widget.NewSeparator(),
		),
		container.NewCenter(closeBtn),
		nil,
		nil,
		container.NewScroll(itemsTable),
	)

	bg := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	dialogContent := container.NewMax(bg, container.NewPadded(content))

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

	title := widget.NewLabelWithStyle(
		"MENU PENJUALAN BARANG",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	search := widget.NewEntry()
	search.SetPlaceHolder("Search...")

	header := container.NewGridWithColumns(3, backBtn, title, container.NewMax(search))

	// Table headers
	headers := []string{"Tgl. Nota", "No. Nota", "Customer"}
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	var selectedRow int = -1

	// Table
	table := widget.NewTable(
		func() (int, int) {
			return len(penjualanData) + 1, len(headers)
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
				text.Refresh()
				return
			}

			bg.FillColor = rowBg
			text.Color = color.Black
			text.TextStyle = fyne.TextStyle{}
			text.TextSize = 13

			if id.Row-1 < len(penjualanData) {
				item := penjualanData[id.Row-1]
				switch id.Col {
				case 0:
					text.Text = item.Header.TglNota
					text.Alignment = fyne.TextAlignCenter
				case 1:
					text.Text = item.Header.NoNota
					text.Alignment = fyne.TextAlignCenter
				case 2:
					text.Text = item.Header.Customer
					text.Alignment = fyne.TextAlignLeading
				}
			}
			text.Refresh()
		},
	)

	table.SetColumnWidth(0, 250)
	table.SetColumnWidth(1, 250)
	table.SetColumnWidth(2, 250)

	// Table selection
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 {
			selectedRow = id.Row - 1
		}
	}

	// Footer
	footer := widget.NewLabelWithStyle(
		"Insert = input data   |   V = View data",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	// Table wrapper
	tableWrapper := container.NewCenter(
		container.NewGridWrap(
			fyne.NewSize(780, 400),
			table,
		),
	)

	// Content
	content := container.NewBorder(header, footer, nil, nil, tableWrapper)
	card := widget.NewCard("", "", content)

	// Refresh function
	refreshTable := func() {
		table.Refresh()
	}

	// Keyboard shortcuts
	w.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		switch k.Name {
		case fyne.KeyInsert:
			showAddPenjualanDialog(w, refreshTable)
		case fyne.KeyV:
			if selectedRow >= 0 && selectedRow < len(penjualanData) {
				showViewPenjualanDialog(w, penjualanData[selectedRow])
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