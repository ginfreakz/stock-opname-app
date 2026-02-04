package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fyne-app/internal/state"
)

func HomePage(w fyne.Window, s *state.Session) fyne.CanvasObject {

	bg := canvas.NewImageFromFile("assets/bg-login.jpg")
	bg.FillMode = canvas.ImageFillStretch

	title := widget.NewLabelWithStyle(
		"Home",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	btnPenjualan := widget.NewButton("Penjualan Barang", func() {
		w.SetContent(PenjualanPage(w, s))
	})
	btnPembelian := widget.NewButton("Pembelian Barang", func() {
		w.SetContent(PembelianPage(w, s))
	})
	btnInventory := widget.NewButton("Inventory / Opname", func() {
		w.SetContent(InventoryPage(w, s))
	})
	btnHapus := widget.NewButton("Hapus Data", func() {})

	logout := widget.NewButton("Logout", func() {
		w.SetContent(LoginPage(w, s))
	})
	logout.Importance = widget.DangerImportance

	separator := canvas.NewLine(color.Gray{Y: 120})
	separator.StrokeWidth = 2

	menu := container.NewVBox(
		title,
		btnPenjualan,
		btnPembelian,
		btnInventory,
		btnHapus,
		separator,
		logout,
	)

	card := widget.NewCard("", "", menu)

	cardContainer := container.NewGridWrap(
		fyne.NewSize(360, 300),
		card,
	)

	return container.NewMax(
		bg,
		container.NewCenter(cardContainer),
	)
}