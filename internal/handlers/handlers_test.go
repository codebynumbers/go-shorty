package handlers

import (
	"testing"
)

func TestGenerateHash(t *testing.T) {
	// test stuff here...
	cases := []struct {
		in, want string
	}{
		{"", "811c9dc5"},
		{"/", "2a0c975e"},
		{"http://www.stuff.com", "654d9cc5"},
	}
	for _, c := range cases {
		got := generateHash(c.in)
		if got != c.want {
			t.Errorf("GenerateHash(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
