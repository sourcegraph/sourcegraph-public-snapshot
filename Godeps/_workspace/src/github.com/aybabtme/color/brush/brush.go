// Package brush provides convenience types to use Brush types as plain strings.
//
// You can use a brush by doing :
//
//   str := "Hello Blue World"
//   blueStr := brush.Blue(str)
//
// ...or simply :
//
//	 blueStr := brush.Blue("I'm Blue")
//
package brush

import (
	"github.com/aybabtme/color"
)

// Black gives black text on a white background. Use it like this:
// 		brush.Black("I'm color Black")
type Black string

func (b Black) String() string {
	return color.NewBrush(color.WhitePaint, color.BlackPaint)(string(b))
}

// White gives white text on a dark gray background. Use it like this:
// 		brush.White("I'm color White")
type White string

func (w White) String() string {
	return color.NewBrush(color.DarkGrayPaint, color.WhitePaint)(string(w))
}

// LightGray gives light gray text on a black background. Use it like this:
// 		brush.LightGray("I'm color LightGray")
type LightGray string

func (l LightGray) String() string {
	return color.NewBrush("", color.LightGrayPaint)(string(l))
}

// Blue gives blue text on a black background. Use it like this:
// 		brush.Blue("I'm color Blue")
type Blue string

func (b Blue) String() string {
	return color.NewBrush("", color.BluePaint)(string(b))
}

// Cyan gives cyan text on a black background. Use it like this:
// 		brush.Cyan("I'm color Cyan")
type Cyan string

func (c Cyan) String() string {
	return color.NewBrush("", color.CyanPaint)(string(c))
}

// Green gives green text on a black background. Use it like this:
// 		brush.Green("I'm color Green")
type Green string

func (g Green) String() string {
	return color.NewBrush("", color.GreenPaint)(string(g))
}

// Purple gives purple text on a black background. Use it like this:
// 		brush.Purple("I'm color Purple")
type Purple string

func (p Purple) String() string {
	return color.NewBrush("", color.PurplePaint)(string(p))
}

// Red gives red text on a black background. Use it like this:
// 		brush.Red("I'm color Red")
type Red string

func (r Red) String() string {
	return color.NewBrush("", color.RedPaint)(string(r))
}

// Yellow gives yellow text on a black background. Use it like this:
// 		brush.Yellow("I'm color Yellow")
type Yellow string

func (y Yellow) String() string {
	return color.NewBrush("", color.YellowPaint)(string(y))
}

// DarkBlue gives dark blue text on a black background. Use it like this:
// 		brush.DarkBlue("I'm color DarkBlue")
type DarkBlue string

func (d DarkBlue) String() string {
	return color.NewBrush("", color.DarkBluePaint)(string(d))
}

// DarkCyan gives dark cyan text on a black background. Use it like this:
// 		brush.DarkCyan("I'm color DarkCyan")
type DarkCyan string

func (d DarkCyan) String() string {
	return color.NewBrush("", color.DarkCyanPaint)(string(d))
}

// DarkGray gives dark gray text on a black background. Use it like this:
// 		brush.DarkGray("I'm color DarkGray")
type DarkGray string

func (d DarkGray) String() string {
	return color.NewBrush("", color.DarkGrayPaint)(string(d))
}

// DarkGreen gives dark green text on a black background. Use it like this:
// 		brush.DarkGreen("I'm color DarkGreen")
type DarkGreen string

func (d DarkGreen) String() string {
	return color.NewBrush("", color.DarkGreenPaint)(string(d))
}

// DarkPurple gives dark purple text on a black background. Use it like this:
// 		brush.DarkPurple("I'm color DarkPurple")
type DarkPurple string

func (d DarkPurple) String() string {
	return color.NewBrush("", color.DarkPurplePaint)(string(d))
}

// DarkRed gives dark red text on a black background. Use it like this:
// 		brush.DarkRed("I'm color DarkRed")
type DarkRed string

func (d DarkRed) String() string {
	return color.NewBrush("", color.DarkRedPaint)(string(d))
}

// DarkYellow gives brown text on a black background. Use it like this:
// 		brush.DarkYellow("I'm color DarkYellow")
type DarkYellow string

func (d DarkYellow) String() string {
	return color.NewBrush("", color.DarkYellowPaint)(string(d))
}
