package ui

import (
	"database/sql"
	
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
	statusLabel := widget.NewLabel("")
	statusLabel.Hide()

	loginBtn := widget.NewButton("Login", func() {
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
	})
	loginBtn.Importance = widget.HighImportance

	header := container.NewVBox(
		icon,
		title,
	)

	form := container.NewVBox(
		header,
		username,
		password,
		statusLabel,
		loginBtn,
	)

	card := widget.NewCard("", "", form)

	cardContainer := container.NewGridWrap(
		fyne.NewSize(360, 280),
		card,
	)

	return container.NewMax(
		bg,
		container.NewCenter(cardContainer),
	)
}