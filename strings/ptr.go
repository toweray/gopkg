package strings

// String2Ptr returns a pointer to the given string.
func String2Ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
