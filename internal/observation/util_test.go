package observation

import "testing"

func TestKebabCase(t *testing.T) {
	testCases := map[string]string{
		"":                             "",
		"SomethingPrettyEasy":          "something-pretty-easy",
		"CodeIntelAPI.GetMonikersByID": "code-intel-api.get-monikers-by-id",
	}

	for input, expectedOutput := range testCases {
		if output := kebabCase(input); output != expectedOutput {
			t.Errorf("unexpected kebab case result. want=%q have=%s", expectedOutput, output)
		}
	}
}
