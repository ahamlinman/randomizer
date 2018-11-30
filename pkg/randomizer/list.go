package randomizer

import (
	"bytes"
	"fmt"
	"strings"
)

func embolden(items []string) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = "*" + item + "*"
	}
	return result
}

func bulletize(items []string) string {
	var buf bytes.Buffer
	for _, item := range items {
		buf.WriteString(fmt.Sprintf("• %s\n", item))
	}
	buf.Truncate(buf.Len() - 1)
	return buf.String()
}

func listify(items []string) string {
	return strings.Join(items, ", ")
}
