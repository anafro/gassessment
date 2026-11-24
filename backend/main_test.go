package main

import "testing"

func TestLine(t *testing.T) {
    for n := 0; n < 100; n++ {
        line := generate()
        if len(line) != 6 {
            t.Fatal("match count")
        }
        for _, m := range line {
            sum := 0.0
            for _, o := range m.Odds {
                if o <= 100 {
                    t.Fatal("invalid odds")
                }
                sum += 100 / float64(o)
            }
            if sum < 1.05 || sum > 1.08 {
                t.Fatalf("margin %f", sum)
            }
        }
    }
}
