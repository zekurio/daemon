package arrayutils

// IndexOf returns the index of the first occurrence of the specified element in the specified slice, or -1 if this slice does not contain the element.
func IndexOf[T comparable](s []T, e T) int {
	for i, v := range s {
		if v == e {
			return i
		}
	}
	return -1
}

// Remove removes an element from a slice by its index, but preserves the order of the slice.
func Remove[T comparable](s []T, i int) []T {
	if i == -1 {
		return s
	}
	newSlice := make([]T, len(s)-1)
	copy(newSlice[:i], s[:i])
	copy(newSlice[i:], s[i+1:])
	return newSlice
}

// RemoveLazy removes an element whose index is unknown from a slice.
// Instead, we use the IndexOf function to find the index of the element to be removed.
func RemoveLazy[T comparable](s []T, e T) []T {
	return Remove(s, IndexOf(s, e))
}

// Add adds an element to a slice at the specified index.
// If the index is -1, the element is added at the end of the slice.
func Add[T comparable](s []T, e T, i int) []T {
	if i == -1 {
		return append(s, e)
	}
	s = append(s, s[len(s)-1])
	copy(s[i+1:], s[i:])
	s[i] = e
	return s
}

// Contains returns true if the specified slice contains the specified element.
func Contains[T comparable](s []T, e T) bool {
	return IndexOf(s, e) != -1
}

// Contained returns a slice containing all the elements of the specified subset that are contained in the specified slice.
func Contained[T comparable](subset, s []T) []T {
	var ct []T
	for _, e := range subset {
		if Contains(s, e) {
			ct = append(ct, e)
		}
	}
	return ct
}

// ContainsAny returns true if the specified slice contains any of the specified elements.
func ContainsAny[T comparable](s []T, elements ...T) bool {
	for _, e := range elements {
		if Contains(s, e) {
			return true
		}
	}
	return false
}
