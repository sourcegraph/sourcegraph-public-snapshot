package main

import "testing"

func Test_applyReplacement(t *testing.T) {
	type args struct {
		fileContent string
		ranges      []codeRange
		replacement string
	}
	tests := []struct {
		name        string
		args        args
		wantNewCode string
		wantErr     bool
	}{
		{
			name: "replace same size symbol",
			args: args{
				replacement: "NewType",
				// ranges are unsorted on purpose, the tested code should handle the sorting
				ranges:      testRanges(),
				fileContent: twoTypesOneLine,
			},
			wantNewCode: twoTypesOneLineReplaced,
			wantErr:     false,
		},
		{
			name: "replace longer size symbol",
			args: args{
				replacement: "NewTypeLonger",
				// ranges are unsorted on purpose, the tested code should handle the sorting
				ranges:      testRanges(),
				fileContent: twoTypesOneLine,
			},
			wantNewCode: twoTypesOneLineReplacedWithLonger,
			wantErr:     false,
		},
		{
			name: "replace shorter size symbol",
			args: args{
				replacement: "New",
				// ranges are unsorted on purpose, the tested code should handle the sorting
				ranges:      testRanges(),
				fileContent: twoTypesOneLine,
			},
			wantNewCode: twoTypesOneLineReplacedWithShorter,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNewCode, err := applyReplacement(tt.args.fileContent, tt.args.ranges, tt.args.replacement)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyReplacement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotNewCode != tt.wantNewCode {
				t.Errorf("applyReplacement() gotNewCode = %v, want %v", gotNewCode, tt.wantNewCode)
			}
		})
	}
}

func testRanges() []codeRange {
	return []codeRange{
		// replacement of 'b' parameter
		{
			start: codeLocation{
				line:      2,
				character: 23,
			},
			end: codeLocation{
				line:      2,
				character: 30,
			},
		},
		// replacement of 'a' parameter
		{
			start: codeLocation{
				line:      2,
				character: 12,
			},
			end: codeLocation{
				line:      2,
				character: 19,
			},
		},
	}
}

const twoTypesOneLine = `
{
	func foo(a OldType, b OldType) {
		...
	}
}
`

const twoTypesOneLineReplaced = `
{
	func foo(a NewType, b NewType) {
		...
	}
}
`

const twoTypesOneLineReplacedWithLonger = `
{
	func foo(a NewTypeLonger, b NewTypeLonger) {
		...
	}
}
`

const twoTypesOneLineReplacedWithShorter = `
{
	func foo(a New, b New) {
		...
	}
}
`
