package main

import (
	"context"
	"fmt"
	"testing"
)

func TestCheckCommandVersion(t *testing.T) {
	mockVersionCmd := func(output string) string {
		return fmt.Sprintf(`echo -e "%s"`, output)
	}

	tests := []struct {
		output     string
		constraint string
		wantErr    bool
	}{
		{"git version 1.2.3", "1.2.3", false},
		{"git version 1.2.3", "> 6.2.3", true},
		{"git version v1.2.3", "1.2.3", false},
		{"git version v1.2.3", "> 6.2.3", true},
		{"git \nversion 1.2.3", "1.2.3", false},
		{"git \nversion 1.2.3", "> 6.2.3", true},
		{"git version 3", "1.2.3", true},
		{"git version foobar", "1.2.3", true},
	}

	// extract user environment general informations
	shellPath, shellConfigPath, err := guessUserShell()
	if err != nil {
		t.Fatalf(err.Error())
	}
	userContext := userContext{shellPath: shellPath, shellConfigPath: shellConfigPath}
	ctx := buildUserContext(userContext, context.Background())

	for _, test := range tests {
		t.Run(fmt.Sprintf("constraint %q against %q", test.constraint, test.output), func(t *testing.T) {
			f := checkCommandOutputVersion(mockVersionCmd(test.output), test.constraint)
			err := f(ctx)

			if test.wantErr && err == nil {
				t.Fatalf("want error but got none")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("want no error but got %q", err)
			}
		})
	}
}
