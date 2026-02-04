package theme

import (
	"image/color"

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