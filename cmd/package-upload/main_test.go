package main

import "testing"

func TestManifestPathKeepsFullVersionedName(t *testing.T) {
	got := manifestPath("dist/app-links-0.1.2-linux-amd64")
	want := "dist/app-links-0.1.2-linux-amd64.manifest.json"
	if got != want {
		t.Fatalf("manifestPath() = %q, want %q", got, want)
	}
}
