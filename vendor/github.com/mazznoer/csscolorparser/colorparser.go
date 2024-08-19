// Package csscolorparser provides function for parsing CSS color string as defined in the W3C's CSS color module level 4.
package csscolorparser

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Inspired by https://github.com/deanm/css-color-parser-js

// R, G, B, A values in the range 0..1
type Color struct {
	R, G, B, A float64
}

// Implement the Go color.Color interface.
func (c Color) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R*c.A*65535 + 0.5)
	g = uint32(c.G*c.A*65535 + 0.5)
	b = uint32(c.B*c.A*65535 + 0.5)
	a = uint32(c.A*65535 + 0.5)
	return
}

// RGBA255 returns R, G, B, A values in the range 0..255
func (c Color) RGBA255() (r, g, b, a uint8) {
	r = uint8(c.R*255 + 0.5)
	g = uint8(c.G*255 + 0.5)
	b = uint8(c.B*255 + 0.5)
	a = uint8(c.A*255 + 0.5)
	return
}

// HexString returns CSS hexadecimal string.
func (c Color) HexString() string {
	r, g, b, a := c.RGBA255()
	if a < 255 {
		return fmt.Sprintf("#%02x%02x%02x%02x", r, g, b, a)
	}
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// RGBString returns CSS RGB string.
func (c Color) RGBString() string {
	r, g, b, _ := c.RGBA255()
	if c.A < 1 {
		return fmt.Sprintf("rgba(%d,%d,%d,%v)", r, g, b, c.A)
	}
	return fmt.Sprintf("rgb(%d,%d,%d)", r, g, b)
}

// Name returns name of this color if its available.
func (c Color) Name() (string, bool) {
	r, g, b, _ := c.RGBA255()
	rgb := [3]uint8{r, g, b}
	for k, v := range namedColors {
		if v == rgb {
			return k, true
		}
	}
	return "", false
}

// Implement the Go TextUnmarshaler interface
func (c *Color) UnmarshalText(text []byte) error {
	col, err := Parse(string(text))
	if err != nil {
		return err
	}
	c.R = col.R
	c.G = col.G
	c.B = col.B
	c.A = col.A
	return nil
}

// Implement the Go TextMarshaler interface
func (c Color) MarshalText() ([]byte, error) {
	return []byte(c.HexString()), nil
}

var black = Color{0, 0, 0, 1}

// Parse parses CSS color string and returns, if successful, a Color.
func Parse(s string) (Color, error) {
	input := s
	s = strings.TrimSpace(strings.ToLower(s))

	if s == "transparent" {
		return Color{0, 0, 0, 0}, nil
	}

	// Predefined name / keyword
	c, ok := namedColors[s]
	if ok {
		return Color{float64(c[0]) / 255, float64(c[1]) / 255, float64(c[2]) / 255, 1}, nil
	}

	// Hexadecimal
	if strings.HasPrefix(s, "#") {
		c, ok := parseHex(s[1:])
		if ok {
			return c, nil
		}
		return black, fmt.Errorf("Invalid hex color, %s", input)
	}

	op := strings.Index(s, "(")

	if (op != -1) && strings.HasSuffix(s, ")") {
		fname := strings.TrimSpace(s[:op])
		alpha := 1.0
		okA := true
		s = s[op+1 : len(s)-1]
		s = strings.ReplaceAll(s, ",", " ")
		s = strings.ReplaceAll(s, "/", " ")
		params := strings.Fields(s)

		if fname == "rgb" || fname == "rgba" {
			if len(params) != 3 && len(params) != 4 {
				return black, fmt.Errorf("%s() format needs 3 or 4 parameters, %s", fname, input)
			}
			r, okR := parsePercentOr255(params[0])
			g, okG := parsePercentOr255(params[1])
			b, okB := parsePercentOr255(params[2])
			if len(params) == 4 {
				alpha, okA = parsePercentOrFloat(params[3])
			}
			if okR && okG && okB && okA {
				return Color{
					clamp0_1(r),
					clamp0_1(g),
					clamp0_1(b),
					clamp0_1(alpha),
				}, nil
			}
			return black, fmt.Errorf("Wrong %s() components, %s", fname, input)

		} else if fname == "hsl" || fname == "hsla" {
			if len(params) != 3 && len(params) != 4 {
				return black, fmt.Errorf("%s() format needs 3 or 4 parameters, %s", fname, input)
			}
			h, okH := parseAngle(params[0])
			s, okS := parsePercentOrFloat(params[1])
			l, okL := parsePercentOrFloat(params[2])
			if len(params) == 4 {
				alpha, okA = parsePercentOrFloat(params[3])
			}
			if okH && okS && okL && okA {
				r, g, b := hslToRgb(normalizeAngle(h), clamp0_1(s), clamp0_1(l))
				return Color{r, g, b, clamp0_1(alpha)}, nil
			}
			return black, fmt.Errorf("Wrong %s() components, %s", fname, input)

		} else if fname == "hwb" || fname == "hwba" {
			if len(params) != 3 && len(params) != 4 {
				return black, fmt.Errorf("hwb() format needs 3 or 4 parameters, %s", input)
			}
			H, okH := parseAngle(params[0])
			W, okW := parsePercentOrFloat(params[1])
			B, okB := parsePercentOrFloat(params[2])
			if len(params) == 4 {
				alpha, okA = parsePercentOrFloat(params[3])
			}
			if okH && okW && okB && okA {
				r, g, b := hwbToRgb(normalizeAngle(H), clamp0_1(W), clamp0_1(B))
				return Color{r, g, b, clamp0_1(alpha)}, nil
			}
			return black, fmt.Errorf("Wrong hwb() components, %s", input)

		} else if fname == "hsv" || fname == "hsva" {
			if len(params) != 3 && len(params) != 4 {
				return black, fmt.Errorf("hsv() format needs 3 or 4 parameters, %s", input)
			}
			h, okH := parseAngle(params[0])
			s, okS := parsePercentOrFloat(params[1])
			v, okV := parsePercentOrFloat(params[2])
			if len(params) == 4 {
				alpha, okA = parsePercentOrFloat(params[3])
			}
			if okH && okS && okV && okA {
				r, g, b := hsvToRgb(normalizeAngle(h), clamp0_1(s), clamp0_1(v))
				return Color{r, g, b, clamp0_1(alpha)}, nil
			}
			return black, fmt.Errorf("Wrong hsv() components, %s", input)
		}
	}

	// RGB hexadecimal format without '#' prefix
	c2, ok2 := parseHex(s)
	if ok2 {
		return c2, nil
	}

	return black, fmt.Errorf("Invalid color format, %s", input)
}

// https://stackoverflow.com/questions/54197913/parse-hex-string-to-image-color

func parseHex(s string) (c Color, ok bool) {
	c.A = 1
	ok = true

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		}
		ok = false
		return 0
	}

	n := len(s)
	if n == 6 || n == 8 {
		c.R = float64(hexToByte(s[0])<<4+hexToByte(s[1])) / 255
		c.G = float64(hexToByte(s[2])<<4+hexToByte(s[3])) / 255
		c.B = float64(hexToByte(s[4])<<4+hexToByte(s[5])) / 255
		if n == 8 {
			c.A = float64(hexToByte(s[6])<<4+hexToByte(s[7])) / 255
		}
	} else if n == 3 || n == 4 {
		c.R = float64(hexToByte(s[0])*17) / 255
		c.G = float64(hexToByte(s[1])*17) / 255
		c.B = float64(hexToByte(s[2])*17) / 255
		if n == 4 {
			c.A = float64(hexToByte(s[3])*17) / 255
		}
	} else {
		ok = false
	}
	return
}

func modulo(x, y float64) float64 {
	return math.Mod(math.Mod(x, y)+y, y)
}

func hueToRgb(n1, n2, h float64) float64 {
	h = modulo(h, 6)
	if h < 1 {
		return n1 + ((n2 - n1) * h)
	}
	if h < 3 {
		return n2
	}
	if h < 4 {
		return n1 + ((n2 - n1) * (4 - h))
	}
	return n1
}

// h = 0..360
// s, l = 0..1
// r, g, b = 0..1
func hslToRgb(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		return l, l, l
	}
	var n2 float64
	if l < 0.5 {
		n2 = l * (1 + s)
	} else {
		n2 = l + s - (l * s)
	}
	n1 := 2*l - n2
	h /= 60
	r = clamp0_1(hueToRgb(n1, n2, h+2))
	g = clamp0_1(hueToRgb(n1, n2, h))
	b = clamp0_1(hueToRgb(n1, n2, h-2))
	return
}

func hwbToRgb(hue, white, black float64) (r, g, b float64) {
	if white+black >= 1 {
		gray := white / (white + black)
		return gray, gray, gray
	}
	r, g, b = hslToRgb(hue, 1, 0.5)
	r = r*(1-white-black) + white
	g = g*(1-white-black) + white
	b = b*(1-white-black) + white
	return
}

func hsvToHsl(H, S, V float64) (h, s, l float64) {
	h = H
	s = S
	l = (2 - S) * V / 2
	if l != 0 {
		if l == 1 {
			s = 0
		} else if l < 0.5 {
			s = S * V / (l * 2)
		} else {
			s = S * V / (2 - l*2)
		}
	}
	return
}

func hsvToRgb(H, S, V float64) (r, g, b float64) {
	h, s, l := hsvToHsl(H, S, V)
	return hslToRgb(h, s, l)
}

func clamp0_1(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

func parseFloat(s string) (float64, bool) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f, err == nil
}

func parsePercentOrFloat(s string) (float64, bool) {
	if strings.HasSuffix(s, "%") {
		f, ok := parseFloat(s[:len(s)-1])
		if ok {
			return f / 100, true
		}
		return 0, false
	}
	return parseFloat(s)
}

func parsePercentOr255(s string) (float64, bool) {
	if strings.HasSuffix(s, "%") {
		f, ok := parseFloat(s[:len(s)-1])
		if ok {
			return f / 100, true
		}
		return 0, false
	}
	f, ok := parseFloat(s)
	if ok {
		return f / 255, true
	}
	return 0, false
}

// Result angle in degrees (not normalized)
func parseAngle(s string) (float64, bool) {
	if strings.HasSuffix(s, "deg") {
		return parseFloat(s[:len(s)-3])
	}
	if strings.HasSuffix(s, "grad") {
		f, ok := parseFloat(s[:len(s)-4])
		if ok {
			return f / 400 * 360, true
		}
		return 0, false
	}
	if strings.HasSuffix(s, "rad") {
		f, ok := parseFloat(s[:len(s)-3])
		if ok {
			return f / math.Pi * 180, true
		}
		return 0, false
	}
	if strings.HasSuffix(s, "turn") {
		f, ok := parseFloat(s[:len(s)-4])
		if ok {
			return f * 360, true
		}
		return 0, false
	}
	return parseFloat(s)
}

func normalizeAngle(t float64) float64 {
	t = math.Mod(t, 360)
	if t < 0 {
		t += 360
	}
	return t
}
