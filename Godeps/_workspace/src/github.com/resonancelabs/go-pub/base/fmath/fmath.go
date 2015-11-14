/*
   A collection of additional math routines, complementing what's in the
   standard math package.

   Most notably this adds float32 functions to simplify code like...

   z := float32( math.Max(float64(x), float64(y)) )

   ...down to:

   z := fmath.Max32(x, y)
*/
package fmath

import "math"

////// float32 /////////////////////////////////////////////////////////////////////////////

func Abs32(a float32) float32 {
	if a < 0 {
		return -a
	} else {
		return a
	}
}

func Min32(a, b float32) float32 {
	if a < b {
		return a
	} else {
		return b
	}
}

func Max32(a, b float32) float32 {
	if a > b {
		return a
	} else {
		return b
	}
}

func Pow32(a, b float32) float32 {
	return float32(math.Pow(float64(a), float64(b)))
}

////// float64 /////////////////////////////////////////////////////////////////////////////

func Lerp(a, b, t float64) float64 {
	return a*(1.0-t) + b*t
}

func Clamp(min, max, v float64) float64 {
	return math.Max(min, math.Min(max, v))
}

func Ease(t float64) float64 {
	return t * t * (3.0 - 2.0*t)
}

func Smoothstep(min, max, t float64) float64 {
	return Ease(Clamp(0.0, 1.0, (t-min)/(max-min)))
}

func Round64(t float64) float64 {
	if t < 0 {
		return math.Ceil(t - 0.5)
	}
	return math.Floor(t + 0.5)
}
