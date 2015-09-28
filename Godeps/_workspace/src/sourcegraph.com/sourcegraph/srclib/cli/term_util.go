package cli

func bold(s string) string {
	return "\x1b[1m" + s + "\x1b[0m"
}

func red(s string) string {
	return "\x1b[31m" + s + "\x1b[0m"
}

func redbg(s string) string {
	return "\x1b[41;37;1m" + s + "\x1b[0m"
}

func green(s string) string {
	return "\x1b[32m" + s + "\x1b[0m"
}

func yellow(s string) string {
	return "\x1b[33m" + s + "\x1b[0m"
}

func blue(s string) string {
	return "\x1b[34m" + s + "\x1b[0m"
}

func cyan(s string) string {
	return "\x1b[36m" + s + "\x1b[0m"
}

func gray(s string) string {
	return "\x1b[30m" + s + "\x1b[0m"
}

func fade(s string) string {
	return "\x1b[30;1m" + s + "\x1b[0m"
}

func invert(s string) string {
	return "\x1b[30m\x1b[47m" + s + "\x1b[0m"
}
