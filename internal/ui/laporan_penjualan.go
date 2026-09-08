package ui

import (
	"fmt"
	"image/color"
	"sort"
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

var laporanMonthOptions = []string{
	"Semua", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember",
}

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

	whiteLabel := func(text string) *canvas.Text {
		t := canvas.NewText(text, color.White)
		t.TextStyle = fyne.TextStyle{Bold: true}
		return t
	}

	// applyFilter is assigned below, once allData/table exist; declared early so
	// the date-picker callbacks (built as part of the header) can already call it.
	var applyFilter func()

	yearSelect := widget.NewSelect([]string{"Semua"}, nil)
	yearSelect.PlaceHolder = "Tahun"

	monthSelect := widget.NewSelect(laporanMonthOptions, nil)
	monthSelect.PlaceHolder = "Bulan"

	var dariDate, sampaiDate string

	dariLabel := whiteLabel("-")
	sampaiLabel := whiteLabel("-")

	dariBtn := widget.NewButtonWithIcon("", theme.CalendarIcon(), func() {
		ShowDatePickerDialog(w, dariDate, func(selectedDate string) {
			dariDate = selectedDate
			dariLabel.Text = selectedDate
			dariLabel.Refresh()
			if applyFilter != nil {
				applyFilter()
			}
		})
	})
	dariBtn.Importance = widget.LowImportance

	sampaiBtn := widget.NewButtonWithIcon("", theme.CalendarIcon(), func() {
		ShowDatePickerDialog(w, sampaiDate, func(selectedDate string) {
			sampaiDate = selectedDate
			sampaiLabel.Text = selectedDate
			sampaiLabel.Refresh()
			if applyFilter != nil {
				applyFilter()
			}
		})
	})
	sampaiBtn.Importance = widget.LowImportance

	clearRangeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		dariDate = ""
		sampaiDate = ""
		dariLabel.Text = "-"
		dariLabel.Refresh()
		sampaiLabel.Text = "-"
		sampaiLabel.Refresh()
		if applyFilter != nil {
			applyFilter()
		}
	})
	clearRangeBtn.Importance = widget.LowImportance

	filterRow := container.NewCenter(container.NewHBox(
		whiteLabel("Tahun:"), yearSelect,
		whiteLabel("Bulan:"), monthSelect,
		whiteLabel("Dari:"), dariLabel, dariBtn,
		whiteLabel("Sampai:"), sampaiLabel, sampaiBtn,
		clearRangeBtn,
	))

	header := container.NewVBox(
		container.NewBorder(nil, nil, backBtn, nil, container.NewCenter(title)),
		filterRow,
	)

	var allData []LaporanRow
	var data []LaporanRow
	var selectedRow int = -1
	var table *widget.Table

	// Fetch the full sales history once; filtering by day/month/year happens client-side
	fetchData := func() {
		now := time.Now()
		startDate := time.Date(2000, 1, 1, 0, 0, 0, 0, now.Location())
		endDate := now.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		reports, err := s.SellRepo.GetDailyReport(startDate, endDate)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Gagal memuat data: %v", err), w)
			return
		}

		allData = nil
		yearSet := map[int]bool{}
		for _, r := range reports {
			allData = append(allData, LaporanRow{
				Date:             r.SellDate,
				DateStr:          r.SellDate.Format("2006-01-02"),
				TransactionCount: fmt.Sprintf("%d", r.TransactionCount),
				TotalAmount:      FormatCurrency(r.TotalAmount),
			})
			yearSet[r.SellDate.Year()] = true
		}

		years := make([]int, 0, len(yearSet))
		for y := range yearSet {
			years = append(years, y)
		}
		if len(years) == 0 {
			years = append(years, now.Year())
		}
		sort.Sort(sort.Reverse(sort.IntSlice(years)))

		yearOptions := []string{"Semua"}
		for _, y := range years {
			yearOptions = append(yearOptions, fmt.Sprintf("%d", y))
		}
		yearSelect.Options = yearOptions
		yearSelect.Refresh()
	}

	applyFilter = func() {
		selectedRow = -1

		year := yearSelect.Selected
		month := monthSelect.Selected

		var startFilter, endFilter time.Time
		hasStart := dariDate != ""
		hasEnd := sampaiDate != ""
		if hasStart {
			startFilter, _ = time.Parse("2006-01-02", dariDate)
		}
		if hasEnd {
			endFilter, _ = time.Parse("2006-01-02", sampaiDate)
		}
		if hasStart && hasEnd && startFilter.After(endFilter) {
			startFilter, endFilter = endFilter, startFilter
		}

		data = nil
		for _, r := range allData {
			if year != "" && year != "Semua" && fmt.Sprintf("%d", r.Date.Year()) != year {
				continue
			}
			if month != "" && month != "Semua" {
				monthNum := 0
				for i, m := range laporanMonthOptions {
					if m == month {
						monthNum = i // index doubles as month number (1=Januari, ...)
						break
					}
				}
				if int(r.Date.Month()) != monthNum {
					continue
				}
			}
			rDate := time.Date(r.Date.Year(), r.Date.Month(), r.Date.Day(), 0, 0, 0, 0, r.Date.Location())
			if hasStart && rDate.Before(startFilter) {
				continue
			}
			if hasEnd && rDate.After(endFilter) {
				continue
			}
			data = append(data, r)
		}

		if table != nil {
			table.Refresh()
		}
	}

	fetchData()
	yearSelect.SetSelected("Semua")
	monthSelect.SetSelected("Semua")
	applyFilter()

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

	yearSelect.OnChanged = func(string) { applyFilter() }
	monthSelect.OnChanged = func(string) { applyFilter() }


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
