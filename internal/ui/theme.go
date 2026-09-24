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
		return color.NRGBA{R: 0x14, G: 0x16, B: 0x19, A: 0xff} // Neutral dark slate

	case theme.ColorNameInputBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		}
		return color.NRGBA{R: 0x1f, G: 0x22, B: 0x27, A: 0xff}

	case theme.ColorNameOverlayBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xf5}
		}
		return color.NRGBA{R: 0x1f, G: 0x22, B: 0x27, A: 0xf5}

	case theme.ColorNamePrimary:
		// Biru profesional terukur
		return color.NRGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0xff}

	case theme.ColorNameSuccess:
		// Hijau sukses
		return color.NRGBA{R: 0x16, G: 0xa3, B: 0x4a, A: 0xff}

	case theme.ColorNameWarning:
		// Kuning/oranye peringatan
		return color.NRGBA{R: 0xd9, G: 0x77, B: 0x06, A: 0xff}

	case theme.ColorNameError:
		// Merah kesalahan
		return color.NRGBA{R: 0xdc, G: 0x26, B: 0x26, A: 0xff}

	case theme.ColorNameForeground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0x0f, G: 0x17, B: 0x2a, A: 0xff} // Slate-900
		}
		return color.NRGBA{R: 0xf1, G: 0xf5, B: 0xf9, A: 0xff} // Slate-100

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
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x40}

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
		return 18.0
	case theme.SizeNameSubHeadingText:
		return 15.0
	case theme.SizeNameCaptionText:
		return 11.5
	case theme.SizeNameScrollBar:
		return 10.0
	default:
		return t.Theme.Size(name)
	}
}
