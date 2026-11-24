package main

import "cmp"

// Between reports whether value is in the inclusive range [minimum, maximum].
// An inverted range contains no values.
func Between[T cmp.Ordered](value, minimum, maximum T) bool {
    return minimum <= value && value <= maximum
}
