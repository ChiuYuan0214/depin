package color

const (
	yellow = "\033[33m"
	green  = "\033[32m"
	blue   = "\033[34m"
	end    = "\033[0m"
)

func Yellow(text string) string {
	return yellow + text + end
}

func Green(text string) string {
	return green + text + end
}

func Blue(text string) string {
	return blue + text + end
}
