package main

import "testing"

func TestServer(t *testing.T) {
	got := root()
	want := "This is a demo."

	if got != want {
		t.Errorf("Got: %q, want: %q", got, want)
	}
}
