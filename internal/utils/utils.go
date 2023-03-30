package utils

import "strings"

// Contains reports whether string is within string array.
func Contains(sa []string, s string) bool {
	s = strings.TrimSpace(s)
	for _, a := range sa {
		if strings.EqualFold(a, s) {
			return true
		}
	}
	return false
}

// EscapeMarkDownV1Text escapes markdownV1 special characters, used in places where we want to send text as-is.
// For example, telegram username with underscores would be italicized if we don't escape it.
// https://core.telegram.org/bots/api#markdown-style
func EscapeMarkDownV1Text(text string) string {
	escSymbols := []string{"_", "*", "`", "["}
	for _, esc := range escSymbols {
		text = strings.Replace(text, esc, "\\"+esc, -1)
	}

	return text
}

// GenHelpMsg construct help message from bot's ReactOn.
func GenHelpMsg(com []string, msg string) string {
	return EscapeMarkDownV1Text(strings.Join(com, ", ")) + " _- " + msg + "_\n"
}
