// Provides some predefines ANSI codes to change foreground/background colors or apply text effects
package colorable

// Bold text effect
func Bold(s string) string {
	return "\x1b[1m" + s + "\x1b[0m"
}

// Underline text effect
func Underline(s string) string {
	return "\x1b[4m" + s + "\x1b[0m"
}

// Red text
func Red(s string) string {
	return "\x1b[31m" + s + "\x1b[0m"
}

// Dark red text
func DarkRed(s string) string {
	return "\x1b[0;31m" + s + "\x1b[0m"
}

// Red background
func Redbg(s string) string {
	return "\x1b[41;37;1m" + s + "\x1b[0m"
}

// Green text
func Green(s string) string {
	return "\x1b[32m" + s + "\x1b[0m"
}

// Green background
func Greenbg(s string) string {
	return "\x1b[42;37;1m" + s + "\x1b[39;49m"
}

// Yellow text
func Yellow(s string) string {
	return "\x1b[33m" + s + "\x1b[0m"
}

// Blue text
func Blue(s string) string {
	return "\x1b[34m" + s + "\x1b[0m"
}

// Cyan text
func Cyan(s string) string {
	return "\x1b[36m" + s + "\x1b[0m"
}

// Gray text
func Gray(s string) string {
	return "\x1b[30m" + s + "\x1b[0m"
}

// Fade text effect
func Fade(s string) string {
	return "\x1b[30;1m" + s + "\x1b[0m"
}

// Inverted text effect
func Invert(s string) string {
	return "\x1b[30m\x1b[47m" + s + "\x1b[0m"
}
