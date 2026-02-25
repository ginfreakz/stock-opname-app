package ui

import (
	"database/sql"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyne-app/internal/state"
)

func LoginPage(w fyne.Window, s *state.Session) fyne.CanvasObject {

	bg := canvas.NewImageFromFile("assets/bg-login.jpg")
	bg.FillMode = canvas.ImageFillStretch

	icon := widget.NewIcon(theme.AccountIcon())

	title := widget.NewLabelWithStyle(
		"Login",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	username := widget.NewEntry()
	username.SetPlaceHolder("Username")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")

	// Status label for showing errors
	statusLabel := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	statusLabel.Hide()

	doLogin := func() {
		if username.Text == "" || password.Text == "" {
			statusLabel.SetText("Username dan Password harus diisi!")
			statusLabel.Show()
			return
		}

		// Authenticate with database
		user, err := s.UserRepo.Authenticate(username.Text, password.Text)
		if err != nil {
			if err == sql.ErrNoRows {
				statusLabel.SetText("Username atau Password salah!")
			} else {
				statusLabel.SetText("Error: " + err.Error())
			}
			statusLabel.Show()
			return
		}

		// Set session data
		s.IsLoggedIn = true
		s.Username = user.Username
		s.User = user

		// Navigate to home page
		w.SetContent(HomePage(w, s))
	}

	username.OnSubmitted = func(string) {
		w.Canvas().Focus(password)
	}

	password.OnSubmitted = func(string) {
		doLogin()
	}

	header := container.NewVBox(
		icon,
		title,
		widget.NewLabel(""), // Add visual spacing
	)

	form := container.NewVBox(
		header,
		container.NewPadded(username),
		container.NewPadded(password),
		statusLabel,
		widget.NewLabel(""), // Add spacing before instruction
	)

	// Create a semi-transparent dark gray rectangle for the panel
	rect := canvas.NewRectangle(color.NRGBA{R: 30, G: 30, B: 30, A: 180})
	rect.CornerRadius = 12
	rect.StrokeColor = color.NRGBA{R: 255, G: 255, B: 255, A: 40} // Subtle white border
	rect.StrokeWidth = 1
	rect.SetMinSize(fyne.NewSize(380, 320))

	// Container for the form with padding
	formContent := container.NewPadded(form)

	// Stack form on top of the background rectangle
	panel := container.NewMax(
		rect,
		formContent,
	)

	// Center the panel in the window
	centeredPanel := container.NewCenter(panel)

	// Combine background and centered panel
	return container.NewMax(
		bg,            // stretched background
		centeredPanel, // floating glass-effect panel
	)
}
