package ui

import (
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

	loginBtn := widget.NewButton("Login", func() {
		if username.Text != "" && password.Text != "" {
			s.Username = username.Text
			w.SetContent(HomePage(w, s))
		}
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
		loginBtn,
	)

	card := widget.NewCard("", "", form)

	cardContainer := container.NewGridWrap(
		fyne.NewSize(360, 240),
		card,
	)

	return container.NewMax(
		bg,
		container.NewCenter(cardContainer),
	)
}
