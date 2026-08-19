package main

import "strings"

func parseWildcard(s string) (is, multipart bool, name string) {
	if s == "*" {
		return true, false, ""
	}

	if s == "**" {
		return true, true, ""
	}

	if strings.HasPrefix(s, "*:") {
		return true, false, s[2:]
	}

	if strings.HasPrefix(s, "**:") {
		return true, true, s[3:]
	}

	return false, false, ""
}
