package cli

const resetColor = "\x1b[0m"

func normal(s string) string { return s }

func bold(s string) string {
	return "\x1b[1m" + s + resetColor
}

func underline(s string) string {
	return "\x1b[4m" + s + resetColor
}

func red(s string) string {
	return "\x1b[31m" + s + resetColor
}

func redbg(s string) string {
	return "\x1b[41;37;1m" + s + resetColor
}

func green(s string) string {
	return "\x1b[32m" + s + resetColor
}

func greenbg(s string) string {
	return "\x1b[42;37;1m" + s + "\x1b[39;49m"
}

func yellow(s string) string {
	return "\x1b[33m" + s + resetColor
}

func blue(s string) string {
	return "\x1b[34m" + s + resetColor
}

func cyan(s string) string {
	return "\x1b[36m" + s + resetColor
}

func gray(s string) string {
	return "\x1b[30m" + s + resetColor
}

func fade(s string) string {
	return "\x1b[30;1m" + s + resetColor
}

func invert(s string) string {
	return "\x1b[30m\x1b[47m" + s + resetColor
}
