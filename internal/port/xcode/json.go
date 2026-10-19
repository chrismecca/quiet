// @(#)quiet / go / internal/port/xcode/json.go / 2026-10-18

package xcode

import (
	"encoding/json"
	"fmt"
	"io"
	"math"

	"github.com/chrismecca/quiet/internal/color"
)

type exactDocument struct {
	Exact float64 `json:"exact"`
}

type hueDocument struct {
	Radians float64 `json:"radians"`
}

type colorDocument struct {
	Chroma    exactDocument `json:"chroma"`
	Gamut     string        `json:"gamut"`
	Hue       hueDocument   `json:"hue"`
	Lightness exactDocument `json:"lightness"`
	Opacity   float64       `json:"opacity"`
}

type backgroundDocument struct {
	CustomColor colorDocument `json:"customColor"`
}

type customDocument struct {
	Custom colorDocument `json:"custom"`
}

type overrideDocument struct {
	Color            *colorDocument `json:"color,omitempty"`
	ColorWithOpacity *colorDocument `json:"colorWithOpacity,omitempty"`
}

type paletteDocument struct {
	ColorOverrides map[string]overrideDocument `json:"colorOverrides"`
	Primary        customDocument              `json:"primary"`
	Secondary      customDocument              `json:"secondary"`
}

type recipeDocument struct {
	Background  backgroundDocument `json:"background"`
	ColorScheme string             `json:"colorScheme"`
	Palette     paletteDocument    `json:"palette"`
}

type themeDocument struct {
	FileVersion int            `json:"fileVersion"`
	Recipe      recipeDocument `json:"recipe"`
}

// EncodeJSON writes an Xcode 27 native .xcworkspacecolortheme document.
func EncodeJSON(
	w io.Writer,
	t Theme,
) error {
	background, err := encodeColor(t.Background)
	if err != nil {
		return fmt.Errorf(
			"background: %w",
			err,
		)
	}

	primary, err := encodeColor(t.Primary)
	if err != nil {
		return fmt.Errorf(
			"primary: %w",
			err,
		)
	}

	secondary, err := encodeColor(t.Secondary)
	if err != nil {
		return fmt.Errorf(
			"secondary: %w",
			err,
		)
	}

	overrides := make(
		map[string]overrideDocument,
		len(t.Overrides),
	)

	for _, override := range t.Overrides {
		if _, exists := overrides[override.Role]; exists {
			return fmt.Errorf(
				"duplicate Xcode role %q",
				override.Role,
			)
		}

		resolved, err := encodeColor(override.Color)
		if err != nil {
			return fmt.Errorf(
				"%s: %w",
				override.Role,
				err,
			)
		}

		switch override.Kind {
		case overrideColor:
			value := resolved

			overrides[override.Role] = overrideDocument{
				Color: &value,
			}

		case overrideColorWithOpacity:
			value := resolved

			overrides[override.Role] = overrideDocument{
				ColorWithOpacity: &value,
			}

		default:
			return fmt.Errorf(
				"%s has unknown Xcode override kind %q",
				override.Role,
				override.Kind,
			)
		}
	}

	doc := themeDocument{
		FileVersion: 1,

		Recipe: recipeDocument{
			Background: backgroundDocument{
				CustomColor: background,
			},

			ColorScheme: t.ColorScheme,

			Palette: paletteDocument{
				ColorOverrides: overrides,

				Primary: customDocument{
					Custom: primary,
				},

				Secondary: customDocument{
					Custom: secondary,
				},
			},
		},
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(doc); err != nil {
		return fmt.Errorf(
			"encode Xcode theme: %w",
			err,
		)
	}

	return nil
}

func encodeColor(
	value color.SRGB,
) (colorDocument, error) {
	if !value.InGamut() {
		return colorDocument{}, fmt.Errorf(
			"color is outside sRGB gamut",
		)
	}

	oklch := value.OKLab().OKLCH()

	hue := oklch.H

	if math.IsNaN(hue) {
		hue = 0
	}

	radians := math.Mod(
		hue*math.Pi/180,
		2*math.Pi,
	)

	if radians < 0 {
		radians += 2 * math.Pi
	}

	return colorDocument{
		Chroma: exactDocument{
			Exact: round12(oklch.C),
		},

		Gamut: "sRGB",

		Hue: hueDocument{
			Radians: round12(radians),
		},

		Lightness: exactDocument{
			Exact: round12(oklch.L),
		},

		Opacity: 1,
	}, nil
}

func round12(
	value float64,
) float64 {
	const scale = 1_000_000_000_000.0

	return math.Round(
		value*scale,
	) / scale
}
