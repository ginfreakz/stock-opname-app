package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowDatePickerDialog opens a calendar date picker dialog
func ShowDatePickerDialog(w fyne.Window, currentDate string, onSelected func(string)) {
	var selectedDate time.Time

	if currentDate != "" {
		parsed, err := time.Parse("2006-01-02", currentDate)
		if err == nil {
			selectedDate = parsed
		} else {
			selectedDate = time.Now()
		}
	} else {
		selectedDate = time.Now()
	}

	// Month/Year header
	monthYear := widget.NewLabel(selectedDate.Format("January 2006"))
	monthYear.Alignment = fyne.TextAlignCenter

	// Navigation buttons
	prevBtn := widget.NewButton("<", nil)
	nextBtn := widget.NewButton(">", nil)

	// Calendar grid container
	calendarGrid := container.NewVBox()

	// Function to render the calendar
	renderCalendar := func(date time.Time) {
		calendarGrid.RemoveAll()
		monthYear.SetText(date.Format("January 2006"))

		// Day headers
		dayHeaders := container.NewGridWithColumns(7)
		for _, day := range []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"} {
			text := canvas.NewText(day, color.White)
			text.TextStyle = fyne.TextStyle{Bold: true}
			text.TextSize = 12
			dayHeaders.Add(text)
		}
		calendarGrid.Add(dayHeaders)

		// Day buttons
		daysGrid := container.NewGridWithColumns(7)
		firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
		lastDay := firstDay.AddDate(0, 1, -1).Day()
		startWeekday := int(firstDay.Weekday())

		// Empty cells for days before month starts
		for i := 0; i < startWeekday; i++ {
			daysGrid.Add(widget.NewLabel(""))
		}

		// Day buttons
		for day := 1; day <= lastDay; day++ {
			d := day // Capture for closure
			monthDate := time.Date(date.Year(), date.Month(), d, 0, 0, 0, 0, time.Local)

			btn := widget.NewButton(fmt.Sprintf("%d", d), func() {
				selectedDate = monthDate
				onSelected(monthDate.Format("2006-01-02"))
			})

			// Only highlight the currently selected date
			if monthDate.Format("2006-01-02") == selectedDate.Format("2006-01-02") {
				btn.Importance = widget.HighImportance
			} else {
				btn.Importance = widget.MediumImportance
			}

			daysGrid.Add(btn)
		}

		calendarGrid.Add(daysGrid)
	}

	// Initial render
	renderCalendar(selectedDate)

	// Navigation handlers
	prevBtn.OnTapped = func() {
		selectedDate = selectedDate.AddDate(0, -1, 0)
		renderCalendar(selectedDate)
	}
	nextBtn.OnTapped = func() {
		selectedDate = selectedDate.AddDate(0, 1, 0)
		renderCalendar(selectedDate)
	}

	// Header with navigation
	header := container.NewBorder(
		nil, nil,
		prevBtn, nextBtn,
		monthYear,
	)

	// Main content
	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		calendarGrid,
	)

	// Dialog
	d := dialog.NewCustomConfirm(
		"Pilih Tanggal",
		"Tutup",
		"",
		content,
		func(bool) {},
		w,
	)
	d.Resize(fyne.NewSize(400, 350))
	d.Show()
}
