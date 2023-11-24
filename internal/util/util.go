package util

// OrString returns the first non-empty string.
func OrString(o ...string) string {
	for _, s := range o {
		if s != "" {
			return s
		}
	}

	return ""
}
