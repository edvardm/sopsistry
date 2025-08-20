package core

import "fmt"

// Option represents a value that may or may not be present
type Option[T any] struct {
	value T
	valid bool
}

// Some creates an Option containing a value
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, valid: true}
}

// None creates an empty Option
func None[T any]() Option[T] {
	return Option[T]{valid: false}
}

// IsSome returns true if the Option contains a value
func (o Option[T]) IsSome() bool {
	return o.valid
}

// IsNone returns true if the Option is empty
func (o Option[T]) IsNone() bool {
	return !o.valid
}

// Unwrap returns the contained value, panics if None
func (o Option[T]) Unwrap() T {
	if !o.valid {
		panic("called Unwrap() on None Option")
	}
	return o.value
}

// UnwrapOr returns the contained value or a default
func (o Option[T]) UnwrapOr(defaultValue T) T {
	if o.valid {
		return o.value
	}
	return defaultValue
}

// UnwrapOrElse returns the contained value or calls a function
func (o Option[T]) UnwrapOrElse(fn func() T) T {
	if o.valid {
		return o.value
	}
	return fn()
}

// Map transforms the contained value if present
func Map[T, U any](o Option[T], fn func(T) U) Option[U] {
	if o.valid {
		return Some(fn(o.value))
	}
	return None[U]()
}

// AndThen chains operations that return Options (flatMap)
func AndThen[T, U any](o Option[T], fn func(T) Option[U]) Option[U] {
	if o.valid {
		return fn(o.value)
	}
	return None[U]()
}

// Result represents a value or an error
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a successful Result
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err creates an error Result
func Err[T any](err error) Result[T] {
	var zero T
	return Result[T]{value: zero, err: err}
}

// IsOk returns true if the Result is successful
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsErr returns true if the Result contains an error
func (r Result[T]) IsErr() bool {
	return r.err != nil
}

// Unwrap returns the value, panics on error
func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic(fmt.Sprintf("called Unwrap() on error Result: %v", r.err))
	}
	return r.value
}

// UnwrapOr returns the value or a default on error
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.err != nil {
		return defaultValue
	}
	return r.value
}

// Error returns the error if present
func (r Result[T]) Error() error {
	return r.err
}

// MapResult transforms the value if successful
func MapResult[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return Ok(fn(r.value))
}

// AndThenResult chains operations that return Results
func AndThenResult[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return fn(r.value)
}
