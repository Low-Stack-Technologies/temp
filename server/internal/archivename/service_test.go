package archivename

import (
	"testing"

	"tech.low-stack.temp/server/internal/db"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "photos.zip", want: "photos"},
		{input: "photos.png", want: "photos"},
		{input: "../my-photos", want: "my-photos"},
		{input: "", want: ""},
	}
	for _, test := range tests {
		got := Normalize(test.input)
		if test.want == "" {
			if got != nil {
				t.Fatalf("Normalize(%q) = %q, want nil", test.input, *got)
			}
			continue
		}
		if got == nil || *got != test.want {
			t.Fatalf("Normalize(%q) = %v, want %q", test.input, got, test.want)
		}
	}
}

func TestFilename(t *testing.T) {
	custom := "photos"
	if got := Filename(&custom, nil); got != "photos.zip" {
		t.Fatalf("custom Filename() = %q, want photos.zip", got)
	}

	files := []db.GroupFile{
		{Filename: "photos-1.jpg"},
		{Filename: "photos-2.png"},
	}
	if got := Filename(nil, files); got != "photos.zip" {
		t.Fatalf("common-prefix Filename() = %q, want photos.zip", got)
	}

	files = []db.GroupFile{{Filename: "one.txt"}, {Filename: "two.txt"}}
	if got := Filename(nil, files); got != "files.zip" {
		t.Fatalf("fallback Filename() = %q, want files.zip", got)
	}
}
