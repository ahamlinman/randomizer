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
	switch len(items) {
	case 0:
		return "nothing" // Though in practice this should never happen...

	case 1:
		return items[0]

	case 2:
		return fmt.Sprintf("%s and %s", items[0], items[1])

	default:
		last := len(items) - 1
		return fmt.Sprintf("%s, and %s", strings.Join(items[:last], ", "), items[last])
		//                    ^ Oxford comma - VERY IMPORTANT
	}
}
