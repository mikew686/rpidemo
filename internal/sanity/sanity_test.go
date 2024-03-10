package sanity

import "testing"

func TestHello(t *testing.T) {
    want := "Working!"
    if got := Sanity(); got != want {
        t.Errorf("Sanity() = %q, want %q", got, want)
    }
}
