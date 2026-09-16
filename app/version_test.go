package app

import (
	"testing"
	"time"
)

func TestFormatVersionKeepsExplicit(t *testing.T) {
	if got := formatVersion("1.2.3"); got != "1.2.3" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatVersionDevUsesDateNightly(t *testing.T) {
	want := time.Now().Format("060102") + "-nightly.0"
	if got := formatVersion("dev"); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got := formatVersion(" "); got != want {
		t.Fatalf("empty: got %q want %q", got, want)
	}
}
