package main

import (
	"log"
	
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	
	"time"

	"fyne-app/internal/config"
	"fyne-app/internal/state"
	"fyne-app/internal/theme"
	"fyne-app/internal/ui"
)

func main() {
	a := app.New()
	a.Settings().SetTheme(&theme.AppTheme{})
	w := a.NewWindow("Program SO")

	// Panggil auto cleanup hapus PDF yang lebih tua dari 3 bulan (~90 hari)
	go ui.AutoCleanupTempFolder(90 * 24 * time.Hour)

	// Initialize database connection
	dbConfig := config.NewDBConfig()
	db, err := config.ConnectDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		dialog.ShowError(err, w)
		return
	}
	defer db.Close()

	// Initialize session with database
	session := state.NewSession(db)

	w.SetContent(ui.LoginPage(w, session))
	w.Resize(fyne.NewSize(500, 380))
	w.ShowAndRun()
}