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
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyne-app/internal/models"
	"fyne-app/internal/state"
	"github.com/google/uuid"
)

type InventoryItem struct {
	ID    uuid.UUID
	Code  string
	Name  string
	Qty   string
	Price string
}

// focusableTable wraps a Table and intercepts key events so keyboard shortcuts
// work even after clicking inside the table.
type focusableTable struct {
	widget.BaseWidget
	table      *widget.Table
	onTypedKey func(*fyne.KeyEvent)
	focused    bool
}

func newFocusableTable(table *widget.Table, onTypedKey func(*fyne.KeyEvent)) *focusableTable {
	ft := &focusableTable{
		table:      table,
		onTypedKey: onTypedKey,
	}
	ft.ExtendBaseWidget(ft)
	return ft
}

func (ft *focusableTable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ft.table)
}

func (ft *focusableTable) FocusGained() { ft.focused = true }
func (ft *focusableTable) FocusLost()   { ft.focused = false }
func (ft *focusableTable) TypedRune(r rune) {}

func (ft *focusableTable) TypedKey(ev *fyne.KeyEvent) {
	if ft.onTypedKey != nil {
		ft.onTypedKey(ev)
	}
}

func (ft *focusableTable) KeyDown(ev *fyne.KeyEvent) {}
func (ft *focusableTable) KeyUp(ev *fyne.KeyEvent)   {}

var _ fyne.Focusable = (*focusableTable)(nil)
var _ desktop.Keyable = (*focusableTable)(nil)

func showAddInventoryDialog(w fyne.Window, s *state.Session, dialogOpen *bool, refreshCallback func()) {

	kode := widget.NewEntry()
	nama := widget.NewEntry()
	qty := widget.NewEntry()
	price := widget.NewEntry()

	form := widget.NewForm(
		widget.NewFormItem("Kode", kode),
		widget.NewFormItem("Nama", nama),
		widget.NewFormItem("Qty", qty),
		widget.NewFormItem("Harga", price),
	)

	bg := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	content := container.NewPadded(form)
	formContent := container.NewMax(bg, content)

	var d dialog.Dialog

	submitBtn := widget.NewButton("Submit", func() {
		if kode.Text == "" || nama.Text == "" || qty.Text == "" || price.Text == "" {
			dialog.ShowError(fmt.Errorf("Semua field harus diisi!"), w)
			return
		}

		qtyVal, err := strconv.ParseFloat(qty.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Qty harus berupa angka!"), w)
			return
		}

		priceVal, err := strconv.ParseFloat(price.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Harga harus berupa angka!"), w)
			return
		}

		item := &models.Item{
			Code:      kode.Text,
			Name:      nama.Text,
			Qty:       qtyVal,
			Price:     priceVal,
			CreatedBy: &s.User.ID,
		}

		err = s.ItemRepo.Create(item)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Gagal menyimpan data: %v", err), w)
			return
		}

		d.Hide()
		*dialogOpen = false

		dialog.ShowInformation("Success", "Data berhasil disimpan!", w)

		if refreshCallback != nil {
			refreshCallback()
		}
	})
	submitBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
		*dialogOpen = false
	})
	cancelBtn.Importance = widget.DangerImportance

	buttons := container.NewGridWithColumns(2, cancelBtn, submitBtn)
	finalContent := container.NewBorder(nil, buttons, nil, nil, formContent)

	dialogContent := container.NewMax(
		canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		container.NewPadded(finalContent),
	)

	d = dialog.NewCustom("Add new data", "", dialogContent, w)
	d.Resize(fyne.NewSize(420, 320))
	d.Show()
}

func showEditInventoryDialog(w fyne.Window, s *state.Session, item InventoryItem, dialogOpen *bool, refreshCallback func()) {

	kode := widget.NewEntry()
	kode.SetText(item.Code)
	kode.Disable()

	nama := widget.NewEntry()
	nama.SetText(item.Name)

	qty := widget.NewEntry()
	qty.SetText(item.Qty)

	price := widget.NewEntry()
	price.SetText(item.Price)

	form := widget.NewForm(
		widget.NewFormItem("Kode", kode),
		widget.NewFormItem("Nama", nama),
		widget.NewFormItem("Qty", qty),
		widget.NewFormItem("Harga", price),
	)

	bg := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	content := container.NewPadded(form)
	formContent := container.NewMax(bg, content)

	var d dialog.Dialog

	submitBtn := widget.NewButton("Submit", func() {
		if nama.Text == "" || qty.Text == "" || price.Text == "" {
			dialog.ShowError(fmt.Errorf("Semua field harus diisi!"), w)
			return
		}

		qtyVal, err := strconv.ParseFloat(qty.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Qty harus berupa angka!"), w)
			return
		}

		priceVal, err := strconv.ParseFloat(price.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Harga harus berupa angka!"), w)
			return
		}

		updatedItem := &models.Item{
			ID:        item.ID,
			Code:      item.Code,
			Name:      nama.Text,
			Qty:       qtyVal,
			Price:     priceVal,
			UpdatedBy: &s.User.ID,
		}

		err = s.ItemRepo.Update(updatedItem)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Gagal mengupdate data: %v", err), w)
			return
		}

		d.Hide()
		*dialogOpen = false

		dialog.ShowInformation("Success", "Data berhasil diupdate!", w)

		if refreshCallback != nil {
			refreshCallback()
		}
	})
	submitBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		d.Hide()
		*dialogOpen = false
	})
	cancelBtn.Importance = widget.DangerImportance

	buttons := container.NewGridWithColumns(2, cancelBtn, submitBtn)
	finalContent := container.NewBorder(nil, buttons, nil, nil, formContent)

	dialogContent := container.NewMax(
		canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		container.NewPadded(finalContent),
	)

	d = dialog.NewCustom("Edit data", "", dialogContent, w)
	d.Resize(fyne.NewSize(420, 320))
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
		"Harga",
	}

	var data []InventoryItem
	var selectedRow int = -1

	dialogOpen := false

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
				ID:    item.ID,
				Code:  item.Code,
				Name:  item.Name,
				Qty:   fmt.Sprintf("%.0f", item.Qty),
				Price: fmt.Sprintf("%.0f", item.Price),
			}
		}
	}

	loadData("")

	// ===== COLORS =====
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}
	selectedBg := color.NRGBA{R: 100, G: 150, B: 255, A: 255}

	// ===== TABLE =====
	table := widget.NewTable(
		func() (int, int) {
			return len(data) + 1, len(headers)
		},
		func() fyne.CanvasObject {
			cellBg := canvas.NewRectangle(color.Transparent)
			text := canvas.NewText("", color.Black)
			text.TextSize = 13
			return container.NewMax(cellBg, text)
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			objects := cell.(*fyne.Container).Objects
			cellBg := objects[0].(*canvas.Rectangle)
			text := objects[1].(*canvas.Text)

			if id.Row == 0 {
				cellBg.FillColor = headerBg
				text.Text = headers[id.Col]
				text.Color = color.White
				text.TextSize = 14
				text.TextStyle = fyne.TextStyle{Bold: true}
				text.Alignment = fyne.TextAlignCenter
				text.Refresh()
				return
			}

			if id.Row-1 == selectedRow {
				cellBg.FillColor = selectedBg
				text.Color = color.White
			} else {
				cellBg.FillColor = rowBg
				text.Color = color.Black
			}

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
					text.Text = item.Price
					text.Alignment = fyne.TextAlignTrailing
				}
			}

			text.Refresh()
		},
	)

	var focusWrapper *focusableTable

	safeFocus := func() {
		if focusWrapper == nil {
			return
		}
		if focusWrapper.Size().Width > 0 && focusWrapper.Size().Height > 0 {
			w.Canvas().Focus(focusWrapper)
		}
	}

	refreshTable := func() {
		selectedRow = -1
		loadData(search.Text)
		table.Refresh()
		safeFocus()
	}

	// Table selection — single click only
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 && id.Row-1 < len(data) {
			selectedRow = id.Row - 1
			table.Refresh()

			// Reclaim focus from the inner table using fyne.Do for thread safety
			time.AfterFunc(50*time.Millisecond, func() {
				fyne.Do(func() {
					safeFocus()
				})
			})
		}
	}

	// ===== COLUMN WIDTH =====
	table.SetColumnWidth(0, 150)
	table.SetColumnWidth(1, 350)
	table.SetColumnWidth(2, 120)
	table.SetColumnWidth(3, 160)

	// Search functionality
	search.OnChanged = func(keyword string) {
		selectedRow = -1
		loadData(keyword)
		table.Refresh()
	}

	// ===== KEY HANDLER =====
	handleKey := func(k *fyne.KeyEvent) {
		if dialogOpen {
			return
		}

		switch k.Name {
		case fyne.KeyInsert:
			dialogOpen = true
			showAddInventoryDialog(w, s, &dialogOpen, refreshTable)
		case fyne.KeyE:
			if selectedRow >= 0 && selectedRow < len(data) {
				dialogOpen = true
				showEditInventoryDialog(w, s, data[selectedRow], &dialogOpen, refreshTable)
			} else {
				dialog.ShowInformation("Info", "Pilih data terlebih dahulu!", w)
			}
		case fyne.KeyDelete:
			if selectedRow >= 0 && selectedRow < len(data) {
				dialogOpen = true
				selectedItem := data[selectedRow]
				dialog.ShowConfirm("Konfirmasi",
					fmt.Sprintf("Apakah Anda yakin ingin menghapus '%s'?", selectedItem.Name),
					func(confirmed bool) {
						if confirmed {
							err := s.ItemRepo.Delete(selectedItem.ID, s.User.ID)
							if err != nil {
								dialogOpen = false
								dialog.ShowError(fmt.Errorf("Gagal menghapus data: %v", err), w)
								return
							}
							dialogOpen = false
							dialog.ShowInformation("Success", "Data berhasil dihapus!", w)
							refreshTable()
						} else {
							dialogOpen = false
						}
					}, w)
			} else {
				dialog.ShowInformation("Info", "Pilih data terlebih dahulu!", w)
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

	// ===== FOCUSABLE WRAPPER =====
	focusWrapper = newFocusableTable(table, handleKey)

	// Canvas-level fallback
	w.Canvas().SetOnTypedKey(handleKey)

	// ===== FOOTER =====
	footer := widget.NewLabelWithStyle(
		"Insert = input data   |   E = Edit data   |   Del = Delete data   |   ↑↓ = Navigate",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	// ===== CENTERED TABLE WRAPPER =====
	tableWrapper := container.NewCenter(
		container.NewGridWrap(
			fyne.NewSize(800, 400),
			focusWrapper,
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

	// Defer initial focus until widget is on canvas (fyne.Do for thread safety)
	time.AfterFunc(150*time.Millisecond, func() {
		fyne.Do(func() {
			safeFocus()
		})
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