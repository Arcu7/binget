// Package transform provides functions to transform sequences.
package transform

import "iter"

// Filter returns a sequence containing only the elements that satisfy the predicate.
func Filter[T any](seq iter.Seq[T], predicate func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if predicate(v) && !yield(v) {
				return
			}
		}
	}
}

// FindBy returns the first element in the sequence that satisfies the predicate.
func FindBy[T any](seq iter.Seq[T], predicate func(T) bool) (T, bool) {
	for v := range seq {
		if predicate(v) {
			return v, true
		}
	}
	var zero T

	return zero, false
}
