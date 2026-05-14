package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type ModernLightTheme struct{}

var _ fyne.Theme = (*ModernLightTheme)(nil)

func (t *ModernLightTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	// Always return light-ish theme for this specific design
	switch n {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0xf4, G: 0xf5, B: 0xf7, A: 0xff} // Background gray
	case theme.ColorNameInputBackground, theme.ColorNameMenuBackground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // Pure white for cards/inputs
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x00, G: 0x52, B: 0xcc, A: 0xff} // Jira-like Blue
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0x17, G: 0x2b, B: 0x4d, A: 0xff} // Dark Blue-Gray text
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x22}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 0xeb, G: 0xec, B: 0xf0, A: 0xff}
	}

	return theme.DefaultTheme().Color(n, theme.VariantLight)
}

func (t *ModernLightTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (t *ModernLightTheme) Font(s fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(s)
}

func (t *ModernLightTheme) Size(n fyne.ThemeSizeName) float32 {
	switch n {
	case theme.SizeNamePadding:
		return 12
	case theme.SizeNameScrollBar:
		return 4
	case theme.SizeNameText:
		return 13
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInlineIcon:
		return 18
	case theme.SizeNameInnerPadding:
		return 8
	}
	return theme.DefaultTheme().Size(n)
}
