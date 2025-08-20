package core

// Find searches for the first element matching the predicate
func Find[T any](slice []T, predicate func(T) bool) Option[T] {
	for _, item := range slice {
		if predicate(item) {
			return Some(item)
		}
	}
	return None[T]()
}

// Filter returns elements matching the predicate
func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Map transforms each element using the provided function
func MapSlice[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, 0, len(slice))
	for _, item := range slice {
		result = append(result, fn(item))
	}
	return result
}

// Reduce combines all elements using the provided function
func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, item := range slice {
		result = fn(result, item)
	}
	return result
}

// Contains checks if any element matches the predicate
func Contains[T any](slice []T, predicate func(T) bool) bool {
	return Find(slice, predicate).IsSome()
}

// Unique removes duplicates using the provided key function
func Unique[T any, K comparable](slice []T, keyFn func(T) K) []T {
	seen := make(map[K]bool)
	result := make([]T, 0, len(slice))

	for _, item := range slice {
		key := keyFn(item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}

	return result
}

// GroupBy groups elements by the key returned by keyFn
func GroupBy[T any, K comparable](slice []T, keyFn func(T) K) map[K][]T {
	groups := make(map[K][]T)

	for _, item := range slice {
		key := keyFn(item)
		groups[key] = append(groups[key], item)
	}

	return groups
}

// Set provides a type-safe set implementation
type Set[T comparable] struct {
	items map[T]struct{}
}

// NewSet creates a new set
func NewSet[T comparable](items ...T) *Set[T] {
	set := &Set[T]{
		items: make(map[T]struct{}),
	}
	for _, item := range items {
		set.Add(item)
	}
	return set
}

// Add adds an item to the set
func (s *Set[T]) Add(item T) {
	s.items[item] = struct{}{}
}

// Contains checks if the set contains an item
func (s *Set[T]) Contains(item T) bool {
	_, exists := s.items[item]
	return exists
}

// Remove removes an item from the set
func (s *Set[T]) Remove(item T) {
	delete(s.items, item)
}

// Size returns the number of items in the set
func (s *Set[T]) Size() int {
	return len(s.items)
}

// ToSlice returns all items as a slice
func (s *Set[T]) ToSlice() []T {
	result := make([]T, 0, len(s.items))
	for item := range s.items {
		result = append(result, item)
	}
	return result
}

// Union returns a new set containing items from both sets
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	for item := range s.items {
		result.Add(item)
	}
	for item := range other.items {
		result.Add(item)
	}
	return result
}

// Intersection returns a new set containing common items
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	for item := range s.items {
		if other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}
