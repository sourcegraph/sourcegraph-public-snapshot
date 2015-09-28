/*
	A collection of integer math functions, similar to that provided for floats in the standard math package
*/
package imath

// Returns the absolute value of a.  If a is the minimum representable integer, the results are undefined.
func Abs(a int) int {
	if a < 0 {
		return -a
	} else {
		return a
	}
}

func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func Max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

// Ensures a <= Clamp(n, a, b) <= b.   Assumes a <= b.
func Clamp(n, a, b int) int {
	if n > b {
		return b
	} else if n < a {
		return a
	} else {
		return n
	}
}

// Returns the absolute value of a.  If a is math.MinInt32, the results are undefined.
func Abs32(a int32) int32 {
	if a < 0 {
		return -a
	} else {
		return a
	}
}

func Min32(a, b int32) int32 {
	if a < b {
		return a
	} else {
		return b
	}
}

func Max32(a, b int32) int32 {
	if a > b {
		return a
	} else {
		return b
	}
}

// Ensures a <= Clamp(n, a, b) <= b.   Assumes a <= b.
func Clamp32(n, a, b int32) int32 {
	if n > b {
		return b
	} else if n < a {
		return a
	} else {
		return n
	}
}

// Returns the absolute value of a.  If a is math.MinInt64, the results are undefined.
func Abs64(a int64) int64 {
	if a < 0 {
		return -a
	} else {
		return a
	}
}

func Min64(a, b int64) int64 {
	if a < b {
		return a
	} else {
		return b
	}
}

func Max64(a, b int64) int64 {
	if a > b {
		return a
	} else {
		return b
	}
}

// Ensures a <= Clamp(n, a, b) <= b.   Assumes a <= b.
func Clamp64(n, a, b int64) int64 {
	if n > b {
		return b
	} else if n < a {
		return a
	} else {
		return n
	}
}
