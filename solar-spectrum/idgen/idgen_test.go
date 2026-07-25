package idgen

import "testing"

func TestNewLength(t *testing.T) {
	if got := len(New("")); got != length {
		t.Errorf("len(New(\"\")) = %d, want %d", got, length)
	}
	if got := len(New("someone@example.com")); got != length {
		t.Errorf("len(New(email)) = %d, want %d", got, length)
	}
}

func TestNewIsUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := New("someone@example.com")
		if seen[id] {
			t.Fatalf("duplicate ID generated: %s", id)
		}
		seen[id] = true
	}
}
