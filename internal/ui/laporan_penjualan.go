package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyne-app/internal/state"
)

type LaporanRow struct {
	Date             time.Time
	DateStr          string
	TransactionCount string
	TotalAmount      string
}

func showLaporanDetailDialog(w fyne.Window, s *state.Session, date time.Time, onClose func()) {
	// Load all sell headers for this date
	headers, err := s.SellRepo.GetByDate(date)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Gagal memuat data: %v", err), w)
		return
	}

	type DetailRow struct {
		ID       string
		NoNota   string
		Customer string
		Total    string
		Status   string
	}

	var rows []DetailRow
	var grandTotal float64
	for _, h := range headers {
		rows = append(rows, DetailRow{
			ID:       h.ID.String(),
			NoNota:   h.SellInvoiceNum,
			Customer: h.CustomerName,
			Total:    FormatCurrency(h.TotalAmount),
			Status:   h.Status,
		})
		if h.Status == "ACTIVE" {
			grandTotal += h.TotalAmount
		}
	}

	dateLabel := canvas.NewText(
		fmt.Sprintf("Tanggal: %s", date.Format("2006-01-02")),
		color.White,
	)
	dateLabel.TextSize = 14
	dateLabel.TextStyle = fyne.TextStyle{Bold: true}

	countLabel := canvas.NewText(
		fmt.Sprintf("Jumlah Transaksi: %d", len(rows)),
		color.White,
	)
	countLabel.TextSize = 13

	totalLabel := canvas.NewText(
		"Grand Total (ACTIVE): "+FormatCurrency(grandTotal),
		color.White,
	)
	totalLabel.TextSize = 13
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}
	totalLabel.Alignment = fyne.TextAlignTrailing

	// Items table
	colHeaders := []string{"No. Nota", "Customer", "Total", "Status"}
	headerBg := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBg := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	detailTable := widget.NewTable(
		func() (int, int) {
			return len(rows) + 1, len(colHeaders)
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			text := canvas.NewText("", color.Black)
			text.TextSize = 13
			text.Alignment = fyne.TextAlignCenter
			return container.NewMax(bg, text)
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			cont := cell.(*fyne.Container)
			bg := cont.Objects[0].(*canvas.Rectangle)
			text := cont.Objects[1].(*canvas.Text)

			if id.Row == 0 {
				bg.FillColor = headerBg
				text.Text = colHeaders[id.Col]
				text.Color = color.White
				text.TextStyle = fyne.TextStyle{Bold: true}
				text.Alignment = fyne.TextAlignCenter
				text.Refresh()
				return
			}

			bg.FillColor = rowBg
			text.Color = color.Black
			text.TextStyle = fyne.TextStyle{}

			if id.Row-1 < len(rows) {
				row := rows[id.Row-1]

				if row.Status == "VOID" {
					text.Color = color.NRGBA{R: 220, G: 50, B: 50, A: 255}
				}

				switch id.Col {
				case 0:
					if row.Status == "VOID" {
						text.Text = row.NoNota + " [VOID]"
					} else {
						text.Text = row.NoNota
					}
					text.Alignment = fyne.TextAlignCenter
				case 1:
					text.Text = row.Customer
					text.Alignment = fyne.TextAlignCenter
				case 2:
					text.Text = row.Total
					text.Alignment = fyne.TextAlignTrailing
				case 3:
					text.Text = row.Status
					text.Alignment = fyne.TextAlignCenter
				}
			}
			text.Refresh()
		},
	)

	detailTable.SetColumnWidth(0, 180)
	detailTable.SetColumnWidth(1, 200)
	detailTable.SetColumnWidth(2, 150)
	detailTable.SetColumnWidth(3, 110)

	var isSubDialogOpen bool

	// Click row to preview nota detail (reuse existing penjualan preview)
	detailTable.OnSelected = func(id widget.TableCellID) {
		if isSubDialogOpen {
			return
		}
		if id.Row > 0 && id.Row-1 < len(headers) {
			isSubDialogOpen = true
			headerData := headers[id.Row-1]
			sell, err := s.SellRepo.GetByID(headerData.ID)
			if err != nil {
				isSubDialogOpen = false
				dialog.ShowError(err, w)
				return
			}
			showPenjualanDialog(w, s, func() { isSubDialogOpen = false }, sell, false)
		}
	}

	var d dialog.Dialog

	closeBtn := widget.NewButton("Tutup", func() {
		d.Hide()
		if onClose != nil {
			onClose()
		}
	})
	closeBtn.Importance = widget.HighImportance

	tableScroll := container.NewScroll(detailTable)
	tableScroll.SetMinSize(fyne.NewSize(0, 250))

	tableSection := container.NewBorder(
		nil,
		container.NewVBox(
			widget.NewSeparator(),
			container.NewHBox(layout.NewSpacer(), totalLabel),
		),
		nil,
		nil,
		tableScroll,
	)

	content := container.NewBorder(
		container.NewVBox(
			container.NewCenter(widget.NewLabelWithStyle(
				"LAPORAN PENJUALAN HARIAN",
				fyne.TextAlignCenter,
				fyne.TextStyle{Bold: true},
			)),
			container.NewCenter(widget.NewLabelWithStyle(fmt.Sprintf("Detail Tanggal: %s", date.Format("2006-01-02")), fyne.TextAlignCenter, fyne.TextStyle{})),
			widget.NewSeparator(),
			container.NewVBox(
				dateLabel,
				countLabel,
			),
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

func LaporanPenjualanPage(w fyne.Window, s *state.Session) fyne.CanvasObject {
	// Background
	bg := canvas.NewImageFromFile("assets/bg-login.jpg")
	bg.FillMode = canvas.ImageFillStretch

	// Header
	backBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		w.SetContent(HomePage(w, s))
	})

	title := canvas.NewText("LAPORAN PENJUALAN HARIAN", color.White)
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 16

	search := widget.NewEntry()
	search.SetPlaceHolder("Search tanggal (YYYY-MM-DD)...")

	header := container.NewGridWithColumns(3, backBtn, container.NewCenter(title), container.NewMax(search))

	var data []LaporanRow
	var selectedRow int = -1
	var table *widget.Table

	loadData := func(keyword string) {
		selectedRow = -1

		// Fetch wide date range (1 year back to today)
		now := time.Now()
		startDate := time.Date(now.Year()-1, now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate := now.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		reports, err := s.SellRepo.GetDailyReport(startDate, endDate)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Gagal memuat data: %v", err), w)
			return
		}

		data = nil
		for _, r := range reports {
			dateStr := r.SellDate.Format("2006-01-02")
			if keyword != "" && !containsCI(dateStr, keyword) {
				continue
			}
			data = append(data, LaporanRow{
				Date:             r.SellDate,
				DateStr:          dateStr,
				TransactionCount: fmt.Sprintf("%d", r.TransactionCount),
				TotalAmount:      FormatCurrency(r.TotalAmount),
			})
		}
	}

	loadData("")

	// Table
	colHeaders := []string{"Tanggal", "Jumlah Transaksi", "Total Penjualan"}
	headerBgColor := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	rowBgColor := color.NRGBA{R: 235, G: 235, B: 235, A: 255}

	table = widget.NewTable(
		func() (int, int) {
			return len(data) + 1, len(colHeaders)
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			text := canvas.NewText("", color.Black)
			text.TextSize = 13
			text.Alignment = fyne.TextAlignCenter
			return container.NewMax(bg, text)
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			cont := cell.(*fyne.Container)
			bg := cont.Objects[0].(*canvas.Rectangle)
			text := cont.Objects[1].(*canvas.Text)

			if id.Row == 0 {
				bg.FillColor = headerBgColor
				text.Text = colHeaders[id.Col]
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
				bg.FillColor = rowBgColor
				text.Color = color.Black
			}

			text.TextStyle = fyne.TextStyle{}
			text.TextSize = 13

			if id.Row-1 < len(data) {
				item := data[id.Row-1]
				switch id.Col {
				case 0:
					text.Text = item.DateStr
					text.Alignment = fyne.TextAlignCenter
				case 1:
					text.Text = item.TransactionCount
					text.Alignment = fyne.TextAlignCenter
				case 2:
					text.Text = item.TotalAmount
					text.Alignment = fyne.TextAlignTrailing
				}
			}
		},
	)

	table.SetColumnWidth(0, 250)
	table.SetColumnWidth(1, 250)
	table.SetColumnWidth(2, 440)

	// Focus helpers
	var focusWrapper *focusableTable
	safeFocus := func() {
		if focusWrapper != nil {
			fyne.Do(func() {
				w.Canvas().Focus(focusWrapper)
			})
		}
	}

	search.OnChanged = func(keyword string) {
		selectedRow = -1
		loadData(keyword)
		table.Refresh()
	}


	var lastDialogTime time.Time
	var isDialogOpen bool

	// Keyboard shortcuts
	handleKey := func(k *fyne.KeyEvent) {
		if time.Since(lastDialogTime) < 500*time.Millisecond || isDialogOpen {
			return
		}

		switch k.Name {
		// Preview detail
		case fyne.KeyV, fyne.KeyReturn:
			lastDialogTime = time.Now()
			if selectedRow >= 0 && selectedRow < len(data) {
				isDialogOpen = true
				showLaporanDetailDialog(w, s, data[selectedRow].Date, func() {
					isDialogOpen = false
					safeFocus()
				})
			} else {
				dialog.ShowInformation("Info", "Pilih tanggal terlebih dahulu!", w)
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

	// Focusable wrapper
	focusWrapper = newFocusableTable(table, handleKey)

	// Canvas-level fallback
	w.Canvas().SetOnTypedKey(handleKey)

	// Table wrapper
	tableWrapper := container.NewCenter(
		container.NewGridWrap(
			fyne.NewSize(950, 480),
			focusWrapper,
		),
	)

	// Footer
	footer := canvas.NewText(
		"[V] View Detail",
		color.White,
	)
	footer.TextStyle = fyne.TextStyle{Italic: true}
	footer.Alignment = fyne.TextAlignCenter

	// Content
	content := container.NewBorder(header, footer, nil, nil, tableWrapper)

	// Panel
	rect := canvas.NewRectangle(color.NRGBA{R: 30, G: 30, B: 30, A: 180})
	rect.CornerRadius = 12
	rect.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 40}
	rect.StrokeWidth = 1
	rect.SetMinSize(fyne.NewSize(1050, 650))

	panel := container.NewMax(
		rect,
		container.NewPadded(content),
	)

	centeredPanel := container.NewCenter(panel)

	// Initial focus
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
