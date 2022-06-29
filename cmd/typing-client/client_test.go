package main

import "testing"

func TestServer(t *testing.T) {
	got := root()
	want := "This is a client."

	if got != want {
		t.Errorf("Got: %q, want: %q", got, want)
	}
}
