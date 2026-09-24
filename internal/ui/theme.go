package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme mendefinisikan tema visual profesional dengan kontras tinggi untuk fotografer.
type CustomTheme struct {
	fyne.Theme
}

// NewCustomTheme membuat instance CustomTheme baru berbasis tema default Fyne.
func NewCustomTheme() fyne.Theme {
	return &CustomTheme{
		Theme: theme.DefaultTheme(),
	}
}

// Color mengembalikan warna semantik dengan kontras yang terukur dan nyaman di mata fotografer.
func (t *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0xf8, G: 0xfa, B: 0xfc, A: 0xff} // Slate-50
		}
		return color.NRGBA{R: 0x12, G: 0x15, B: 0x1a, A: 0xff} // Slate-950 deep

	case theme.ColorNameInputBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		}
		return color.NRGBA{R: 0x1a, G: 0x1e, B: 0x24, A: 0xff} // Input dark slate

	case theme.ColorNameOverlayBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xf5}
		}
		return color.NRGBA{R: 0x1a, G: 0x1e, B: 0x24, A: 0xf5}

	case theme.ColorNameButton:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0xe2, G: 0xe8, B: 0xf0, A: 0xff} // Slate-200
		}
		return color.NRGBA{R: 0x24, G: 0x2a, B: 0x33, A: 0xff} // Button slate

	case theme.ColorNamePrimary:
		// Biru modern terukur (vibrant blue)
		return color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0xff}

	case theme.ColorNameSuccess:
		// Hijau sukses (emerald)
		return color.NRGBA{R: 0x22, G: 0xc5, B: 0x5e, A: 0xff}

	case theme.ColorNameWarning:
		// Amber peringatan
		return color.NRGBA{R: 0xf5, G: 0x9e, B: 0x0b, A: 0xff}

	case theme.ColorNameError:
		// Crimson kesalahan
		return color.NRGBA{R: 0xef, G: 0x44, B: 0x44, A: 0xff}

	case theme.ColorNameForeground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0x0f, G: 0x17, B: 0x2a, A: 0xff} // Slate-900
		}
		return color.NRGBA{R: 0xf8, G: 0xfa, B: 0xfc, A: 0xff} // Slate-50 crisp white

	case theme.ColorNamePlaceHolder:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0x64, G: 0x74, B: 0x8b, A: 0xff} // Slate-500
		}
		return color.NRGBA{R: 0x94, G: 0xa3, B: 0xb8, A: 0xff} // Slate-400

	case theme.ColorNameScrollBar:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0x94, G: 0xa3, B: 0xb8, A: 0x80}
		}
		return color.NRGBA{R: 0x47, G: 0x55, B: 0x69, A: 0x80}

	case theme.ColorNameShadow:
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x35}

	default:
		return t.Theme.Color(name, variant)
	}
}

// Size mengatur skala ukuran padding dan teks yang proporsional.
func (t *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6.0
	case theme.SizeNameInnerPadding:
		return 8.0
	case theme.SizeNameText:
		return 13.0
	case theme.SizeNameHeadingText:
		return 17.0
	case theme.SizeNameSubHeadingText:
		return 14.5
	case theme.SizeNameCaptionText:
		return 11.5
	case theme.SizeNameScrollBar:
		return 10.0
	default:
		return t.Theme.Size(name)
	}
}
