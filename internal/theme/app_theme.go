package theme

import (
	"image/color"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type AppTheme struct{}

func (AppTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameInputBackground:
		// Light gray instead of pure white
		return color.NRGBA{R: 245, G: 245, B: 245, A: 255}

	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 180, G: 180, B: 180, A: 255}

	// Text colors for input and labels
	case theme.ColorNameForeground:
		// Dark text for light backgrounds
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255}

	case theme.ColorNamePlaceHolder:
		// Gray placeholder text
		return color.NRGBA{R: 128, G: 128, B: 128, A: 255}

	// Background colors
	case theme.ColorNameBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 240}

	// Button colors - Light background for menu buttons
	case theme.ColorNameButton:
		return color.NRGBA{R: 245, G: 245, B: 245, A: 255}

	case theme.ColorNameMenuBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	// Primary color — used by Select widget focus, HighImportance buttons, etc.
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 55, G: 90, B: 220, A: 255}

	// Hover and interaction states - Erase mouse over (hover) effects across the entire app
	case theme.ColorNameHover:
		// Keep select dropdown background solid so it does not turn transparent on hover
		var pcs [8]uintptr
		n := runtime.Callers(2, pcs[:])
		frames := runtime.CallersFrames(pcs[:n])
		for {
			frame, more := frames.Next()
			if strings.Contains(strings.ToLower(frame.Function), "select") {
				return color.NRGBA{R: 245, G: 245, B: 245, A: 255}
			}
			if !more {
				break
			}
		}
		// For buttons and all other widgets, return transparent so original background is preserved
		return color.Transparent

	case theme.ColorNamePressed:
		return color.NRGBA{R: 218, G: 226, B: 240, A: 255}

	case theme.ColorNameFocus:
		return color.NRGBA{R: 55, G: 90, B: 220, A: 120}

	case theme.ColorNameSelection:
		return color.NRGBA{R: 55, G: 90, B: 220, A: 200}
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (AppTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (AppTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (AppTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
