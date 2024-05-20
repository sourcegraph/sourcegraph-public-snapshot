package diagram

import (
	"oss.terrastruct.com/d2/d2target"
	"oss.terrastruct.com/d2/d2themes"
)

// WithSketch sets whether the render should use a hand-drawn aesthetic
// https://d2lang.com/tour/sketch/
//
// If not set defaults to `false`
func WithSketch(sketch bool) func(*diagram) {
	return func(d *diagram) {
		d.renderOpts.Sketch = &sketch
	}
}

// WithTheme sets the theme to use for the render
// Themes are available from the `d2themescatalog` package
//
// If not set defaults to `d2themescatalog.NeutralDefault`
func WithTheme(theme d2themes.Theme) func(*diagram) {
	return func(d *diagram) {
		d.renderOpts.ThemeID = &theme.ID
	}
}

// WithThemeOverrides allows colors from the theme to be overrided
func WithThemeOverrides(themeOverrides d2target.ThemeOverrides) func(*diagram) {
	return func(d *diagram) {
		d.renderOpts.ThemeOverrides = &themeOverrides
	}
}

// WithPadding sets how much padding should be used around the diagram
//
// If not set defaults to `100`
func WithPadding(padding int64) func(*diagram) {
	return func(d *diagram) {
		d.renderOpts.Pad = &padding
	}
}
