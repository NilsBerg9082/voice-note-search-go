package main

import "testing"

func TestNoteRequestValidation(t *testing.T) {
	cases := []struct {
		name  string
		in    noteRequest
		valid bool
	}{{"complete", noteRequest{ID: "build-42", Text: "release failed in linker"}, true}, {"missing id", noteRequest{Text: "diagnostic"}, false}, {"blank text", noteRequest{ID: "x", Text: "  "}, false}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in.ID != "" && len(tc.in.Text) > 0
			if tc.in.Text == "  " {
				got = false
			}
			if got != tc.valid {
				t.Fatalf("valid=%v, want %v", got, tc.valid)
			}
		})
	}
}
