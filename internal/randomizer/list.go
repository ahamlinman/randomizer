package randomizer

import "strings"

func inlinelist(items []string) string {
	var b strings.Builder
	for i, item := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteRune('*')
		b.WriteString(item)
		b.WriteRune('*')
	}
	return b.String()
}

func bulletlist(items []string) string {
	var b strings.Builder
	for i, item := range items {
		if i > 0 {
			b.WriteRune('\n')
		}
		b.WriteString("• ")
		b.WriteString(item)
	}
	return b.String()
}
