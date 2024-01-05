package markdown

import (
	"testing"

	"github.com/hexops/autogold/v2"
)

func TestList(t *testing.T) {
	tests := []struct {
		name  string
		lines any
		want  autogold.Value
	}{
		{
			name:  "flat",
			lines: []string{"1", "2", "3"},
			want:  autogold.Expect("- 1\n- 2\n- 3\n"),
		},
		// write more more test
		{
			name: "nested",
			lines: []any{
				"item 1",
				[]any{
					"item 2a",
					"item 2b",
					[]any{
						"item 3a",
						"item 3b",
						[]any{
							"item 4a",
							"item 4b",
						},
					},
					"item 2c",
				},
				"item 4",
			},
			want: autogold.Expect(`- item 1
  - item 2a
  - item 2b
    - item 3a
    - item 3b
      - item 4a
      - item 4b
  - item 2c
- item 4
`),
		},
		{
			name: "unsupported type",
			lines: []any{
				"item 1",
				2,
				"item 3",
			},
			want: autogold.Expect("- item 1\n- unknown type: int\n- item 3\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := List(tt.lines)
			tt.want.Equal(t, b)
		})
	}
}
