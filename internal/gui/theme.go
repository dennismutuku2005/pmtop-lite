package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type ModernLightTheme struct{}

var _ fyne.Theme = (*ModernLightTheme)(nil)

func (t *ModernLightTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if v == theme.VariantDark {
		switch n {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 0x0f, G: 0x17, B: 0x2a, A: 0xff} // Slate-900
		case theme.ColorNameInputBackground:
			return color.NRGBA{R: 0x1e, G: 0x29, B: 0x3b, A: 0xff} // Slate-800
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 0xfe, G: 0x4a, B: 0x16, A: 0xff} // Brand Orange #fe4a16
		case theme.ColorNameForeground:
			return color.NRGBA{R: 0xf8, G: 0xfa, B: 0xfc, A: 0xff} // Slate-50
		}
	}

	switch n {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0xf8, G: 0xfa, B: 0xfc, A: 0xff} // Slate-50
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // White
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0xfe, G: 0x4a, B: 0x16, A: 0xff} // Brand Orange #fe4a16
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0x0f, G: 0x17, B: 0x2a, A: 0xff} // Slate-900
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x11}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 0xe2, G: 0xe8, B: 0xf0, A: 0xff} // Slate-200
	}

	return theme.DefaultTheme().Color(n, v)
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
		return 16
	case theme.SizeNameScrollBar:
		return 6
	case theme.SizeNameText:
		return 14
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameInnerPadding:
		return 10
	}
	return theme.DefaultTheme().Size(n)
}

