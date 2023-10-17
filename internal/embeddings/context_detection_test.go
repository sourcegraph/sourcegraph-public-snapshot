package embeddings

import (
	"testing"
)

func TestIsContextRequiredForChatQuery(t *testing.T) {
	cases := []struct {
		query string
		want  bool
	}{
		{
			query: "this answer looks incorrect",
			want:  false,
		},
		{
			query: "that doesn’t seem right",
			want:  false,
		},
		{
			query: "I don't understand what you're saying",
			want:  false,
		},
		{
			query: "I don't think that's right",
			want:  false,
		},
		{
			query: "explain that in more detail",
			want:  false,
		},
		{
			query: "are you sure??",
			want:  false,
		},
		{
			query: "what directory contains the cody plugin",
			want:  true,
		},
		{
			query: "Is crewjam/saml used anywhere?",
			want:  true,
		},
		{
			query: "are sub-repo permissions respected in embeddings?",
			want:  true,
		},
		{
			query: "What is BrandLogo",
			want:  true,
		},
		{
			query: "please correct the selected code",
			want:  true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.query, func(t *testing.T) {
			got := IsContextRequiredForChatQuery(tt.query)
			if got != tt.want {
				t.Fatalf("expected context required to be %t but was %t", tt.want, got)
			}
		})
	}
}
