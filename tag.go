package depin

import (
	"strings"
)

func parseTag(tag string) (parsedTag Tag) {
	parts := strings.Split(tag, ":")
	for index, part := range parts {
		if part == "scope" && index < len(parts)-1 {
			parsedTag.scope = scopeType(parts[index+1])
		}
	}
	return
}
