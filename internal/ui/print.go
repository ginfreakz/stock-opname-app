package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"io/ioutil"
	"time"

	"fyne-app/internal/models"

	"github.com/go-pdf/fpdf"
)

// AutoCleanupTempFolder runs asynchronously to delete files in the temp folder older than a certain duration (e.g. 1 hour).
func AutoCleanupTempFolder(maxAge time.Duration) {
	tempDir := "temp"
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return // Silently ignore if folder doesn't exist
	}

	cutoff := time.Now().Add(-maxAge)
	for _, f := range files {
		if !f.IsDir() && f.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(tempDir, f.Name()))
		}
	}
}

// PrintNotaPenjualan generates a PDF receipt for the given SellFull data
// and automatically opens it.
func PrintNotaPenjualan(header models.SellHeader, items []DisplayItem) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(190, 10, "NOTA PENJUALAN", "", 1, "C", false, 0, "")

	if header.Status == "VOID" {
		pdf.SetTextColor(255, 0, 0)
		pdf.SetFont("Arial", "B", 24)
		pdf.CellFormat(190, 10, "** VOID **", "", 1, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0) // Reset to black
		pdf.Ln(2)
	} else {
		pdf.Ln(5)
	}

	// Header Info
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(30, 6, "No. Nota", "", 0, "L", false, 0, "")
	pdf.CellFormat(5, 6, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(100, 6, header.SellInvoiceNum, "", 1, "L", false, 0, "")

	pdf.CellFormat(30, 6, "Tanggal", "", 0, "L", false, 0, "")
	pdf.CellFormat(5, 6, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(100, 6, header.SellDate.Format("02-01-2006"), "", 1, "L", false, 0, "")

	pdf.CellFormat(30, 6, "Customer", "", 0, "L", false, 0, "")
	pdf.CellFormat(5, 6, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(100, 6, header.CustomerName, "", 1, "L", false, 0, "")

	pdf.Ln(5)

	// Table Header
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(15, 8, "No", "1", 0, "C", false, 0, "")
	pdf.CellFormat(75, 8, "Nama Barang", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 8, "Qty", "1", 0, "C", false, 0, "")
	pdf.CellFormat(35, 8, "Harga", "1", 0, "C", false, 0, "")
	pdf.CellFormat(45, 8, "Total", "1", 1, "C", false, 0, "")

	// Table Body
	pdf.SetFont("Arial", "", 10)
	for i, item := range items {
		pdf.CellFormat(15, 8, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(75, 8, fmt.Sprintf("%s - %s", item.Code, item.Name), "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 8, item.Qty, "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 8, item.Price, "1", 0, "R", false, 0, "")
		pdf.CellFormat(45, 8, item.Total, "1", 1, "R", false, 0, "")
	}

	// Grand Total
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(145, 8, "Total", "1", 0, "R", false, 0, "")
	pdf.CellFormat(45, 8, FormatCurrency(header.TotalAmount), "1", 1, "R", false, 0, "")

	// Create temp directory if it doesn't exist
	tempDir := "temp"
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return fmt.Errorf("gagal membuat folder temp: %v", err)
	}

	// Save file
	fileName := filepath.Join(tempDir, fmt.Sprintf("Nota_Penjualan_%s.pdf", header.SellInvoiceNum))
	err := pdf.OutputFileAndClose(fileName)
	if err != nil {
		return err
	}

	return openFile(fileName)
}

func openFile(fileName string) error {
	var err error
	switch runtime.GOOS {
	case "windows":
		// Must use absolute path on windows or cmd won't find it sometimes if args are weird,
		// but simple wrapper is enough if in CWD.
		absPath, _ := filepath.Abs(fileName)
		err = exec.Command("cmd", "/c", "start", "", absPath).Start()
	case "darwin":
		err = exec.Command("open", fileName).Start()
	default: // "linux", "freebsd", "openbsd", "netbsd"
		err = exec.Command("xdg-open", fileName).Start()
	}
	return err
}
