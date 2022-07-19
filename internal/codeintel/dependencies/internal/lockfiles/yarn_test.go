package lockfiles

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParsePackageLocator(t *testing.T) {
	tests := []struct {
		v1              bool
		line            string
		wantName        string
		wantConstraints []constraint
	}{
		//
		// yarn.lock v1
		//
		{
			v1:              true,
			line:            `"@babel/helper-validator-identifier@^7.16.7":`,
			wantName:        "@babel/helper-validator-identifier",
			wantConstraints: []constraint{{Version: "^7.16.7"}},
		},
		{
			v1:              true,
			line:            `"@eslint/eslintrc@~1.3.0":`,
			wantName:        "@eslint/eslintrc",
			wantConstraints: []constraint{{Version: "~1.3.0"}},
		},
		{
			v1:              true,
			line:            `acorn-jsx@^5.3.2:`,
			wantName:        "acorn-jsx",
			wantConstraints: []constraint{{Version: "^5.3.2"}},
		},
		{
			v1:              true,
			line:            `acorn@^8.7.1:`,
			wantName:        "acorn",
			wantConstraints: []constraint{{Version: "^8.7.1"}},
		},
		{
			v1:              true,
			line:            `ajv@^6.10.0, ajv@^6.12.4:`,
			wantName:        "ajv",
			wantConstraints: []constraint{{Version: "^6.10.0"}, {Version: "^6.12.4"}},
		},
		{
			v1:              true,
			line:            `"@types/istanbul-lib-coverage@*", "@types/istanbul-lib-coverage@^2.0.0":`,
			wantName:        "@types/istanbul-lib-coverage",
			wantConstraints: []constraint{{Version: "*"}, {Version: "^2.0.0"}},
		},
		{
			v1:              true,
			line:            `"@types/istanbul-lib-report@*":`,
			wantName:        "@types/istanbul-lib-report",
			wantConstraints: []constraint{{Version: "*"}},
		},
		{
			v1:              true,
			line:            `"safer-buffer@>= 2.1.2 < 3.0.0":`,
			wantName:        "safer-buffer",
			wantConstraints: []constraint{{Version: ">= 2.1.2 < 3.0.0"}},
		},
		{
			v1:              true,
			line:            `"@types/node@14.x || 15.x":`,
			wantName:        "@types/node",
			wantConstraints: []constraint{{Version: "14.x || 15.x"}},
		},
		//
		// yarn.lock v2
		//
		{
			v1:              false,
			line:            `"commander@npm:^6.1.0":`,
			wantName:        "commander",
			wantConstraints: []constraint{{Version: "^6.1.0", Protocol: "npm"}},
		},
		{
			v1:              false,
			line:            `"concat-map@npm:0.0.1":`,
			wantName:        "concat-map",
			wantConstraints: []constraint{{Version: "0.0.1", Protocol: "npm"}},
		},
		{
			v1:              false,
			line:            `"console-control-strings@npm:^1.0.0, console-control-strings@npm:~1.1.0":`,
			wantName:        "console-control-strings",
			wantConstraints: []constraint{{Version: "^1.0.0", Protocol: "npm"}, {Version: "~1.1.0", Protocol: "npm"}},
		},
		{
			v1:              false,
			line:            `"core-util-is@npm:1.0.2, core-util-is@npm:~1.0.0":`,
			wantName:        "core-util-is",
			wantConstraints: []constraint{{Version: "1.0.2", Protocol: "npm"}, {Version: "~1.0.0", Protocol: "npm"}},
		},
	}

	for i, tt := range tests {
		haveName, haveProtocols, err := parsePackageLocatorLine(tt.line, tt.v1)
		if err != nil {
			t.Errorf("test %d. error: %s", i, err)
			continue
		}

		if haveName != tt.wantName {
			t.Errorf("test %d. wrong name. want=%q, have=%q", i, tt.wantName, haveName)
			continue
		}

		if d := cmp.Diff(tt.wantConstraints, haveProtocols); d != "" {
			t.Errorf("test %d. wrong protocols. diff=%s", i, d)
		}
	}
}
