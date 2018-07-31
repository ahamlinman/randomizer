package randomizer

// Selector is a type for functions that, given an integer n > 0, return a
// random integer in the range [0, n).
type Selector func(n int) int

// PickString uses the Selector to select and return a string from the provided
// slice.
func (s Selector) PickString(options []string) string {
	idx := s(len(options))
	return options[idx]
}
