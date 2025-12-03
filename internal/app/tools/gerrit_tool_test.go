package tools

import "testing"

func TestStripXSSI(t *testing.T) {
    b := []byte(")]}'\n{\"a\":1}")
    out := stripXSSI(b)
    if string(out) != "{\"a\":1}" { t.Fatalf("stripXSSI failed: %s", string(out)) }
}
