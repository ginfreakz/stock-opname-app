package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne-app/internal/theme"
	"fyne-app/internal/state"
	"fyne-app/internal/ui"
)

func main() {
	a := app.New()
	a.Settings().SetTheme(&theme.AppTheme{})
	w := a.NewWindow("Program SO")

	session := &state.Session{}

	w.SetContent(ui.LoginPage(w, session))
	w.Resize(fyne.NewSize(500, 380))
	w.ShowAndRun()
}
