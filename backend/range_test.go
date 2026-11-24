package main

import "testing"

func TestBetween(t *testing.T) {
    tests := []struct {
        name                    string
        value, minimum, maximum int64
        want                    bool
    }{
        {"below", 0, 1, 100, false},
        {"lower boundary", 1, 1, 100, true},
        {"inside", 50, 1, 100, true},
        {"upper boundary", 100, 1, 100, true},
        {"above", 101, 1, 100, false},
        {"single value", 1, 1, 1, true},
        {"inverted range", 50, 100, 1, false},
        {"negative range", -5, -10, -1, true},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            if got := Between(test.value, test.minimum, test.maximum); got != test.want {
                t.Fatalf("Between(%d, %d, %d) = %t, want %t", test.value, test.minimum, test.maximum, got, test.want)
            }
        })
    }
}
