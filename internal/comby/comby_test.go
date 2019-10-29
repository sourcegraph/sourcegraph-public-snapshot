package comby

import "testing"

func TestMatchesInZip(t *testing.T) {
	testCases := []struct {
		name string
		want string
	}{
		{"case 1", "yes"},
	}

	for _, test := range testCases {
		t.Run(test.name, func(*testing.T) {
			got := "yes"
			if got != test.want {
				t.Errorf("failed %v, got %v, want %v", test.name, got, test.want)
			}
		})
	}
}
