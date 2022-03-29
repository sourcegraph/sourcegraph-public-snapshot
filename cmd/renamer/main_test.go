package main

import (
	"io/ioutil"
	"os"
	"testing"
)

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
				ranges:      testRangesTwoArgs(),
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
				ranges:      testRangesTwoArgs(),
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
				ranges:      testRangesTwoArgs(),
				fileContent: twoTypesOneLine,
			},
			wantNewCode: twoTypesOneLineReplacedWithShorter,
			wantErr:     false,
		},
		{
			name: "replace three inline function args",
			args: args{
				replacement: "Replaced",
				// ranges are unsorted on purpose, the tested code should handle the sorting
				ranges:      testRangesThreeArgs(),
				fileContent: threeTypesOneLine,
			},
			wantNewCode: threeTypesOneLineReplaced,
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

func testRangesTwoArgs() []codeRange {
	return []codeRange{
		// replacement of 'b' parameter
		{
			start: codeLocation{
				line:      3,
				character: 22,
			},
			end: codeLocation{
				line:      3,
				character: 29,
			},
		},
		// replacement of 'a' parameter
		{
			start: codeLocation{
				line:      3,
				character: 11,
			},
			end: codeLocation{
				line:      3,
				character: 18,
			},
		},
	}
}

func testRangesThreeArgs() []codeRange {
	ranges := testRangesTwoArgs()
	ranges = append(ranges, codeRange{
		start: codeLocation{
			line:      3,
			character: 33,
		},
		end: codeLocation{
			line:      3,
			character: 40,
		},
	})
	return ranges
}

const twoTypesOneLine = `
package main

func foo(a OldType, b OldType) {
	println("hello")
}

type OldType = string
type NewType = string
type New = string
`

const threeTypesOneLine = `
package main

func foo(a OldType, b OldType, c OldType) {
	println("hello")
}

type Replaced = string
`

const threeTypesOneLineReplaced = `
package main

func foo(a Replaced, b Replaced, c Replaced) {
	println("hello")
}

type Replaced = string
`

const twoTypesOneLineReplaced = `
package main

func foo(a NewType, b NewType) {
	println("hello")
}

type OldType = string
type NewType = string
type New = string
`

const twoTypesOneLineReplacedWithLonger = `
package main

func foo(a NewTypeLonger, b NewTypeLonger) {
	println("hello")
}

type OldType = string
type NewType = string
type New = string
`

const twoTypesOneLineReplacedWithShorter = `
package main

func foo(a New, b New) {
	println("hello")
}

type OldType = string
type NewType = string
type New = string
`

func Test_writeReplacement(t *testing.T) {
	type args struct {
		ranges      map[string][]codeRange
		replacement string
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
		expectedContent map[string]string
	}{
		{
			name: "replaces stuff in a single file",
			args: args{
				ranges: map[string][]codeRange{
					"sample_file.go": testRangesTwoArgs(),
				},
				replacement: "NewType",
			},
			expectedContent: map[string]string{
				"sample_file.go": twoTypesOneLineReplaced,
			},
		},
		{
			name: "replaces stuff in a multiple files",
			args: args{
				ranges: map[string][]codeRange{
					"sample_file.go":  testRangesTwoArgs(),
					"sample_file2.go": testRangesTwoArgs(),
				},
				replacement: "NewType",
			},
			expectedContent: map[string]string{
				"sample_file.go":  twoTypesOneLineReplaced,
				"sample_file2.go": twoTypesOneLineReplaced,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wd, err := os.Getwd()
			if err != nil {
				t.Fatal("couldn't get working dir")
			}

			dir, err := ioutil.TempDir(wd, "testing*")
			if err != nil {
				t.Fatal("couldn't create tmp dir here boo")
			}

			err = ioutil.WriteFile(dir+"/sample_file.go", []byte(twoTypesOneLine), 0777)
			if err != nil {
				t.Fatalf("couldn't create the sample file in tmp dir: %s", err)
			}
			err = ioutil.WriteFile(dir+"/sample_file2.go", []byte(twoTypesOneLine), 0777)
			if err != nil {
				t.Fatalf("couldn't create the sample file in tmp dir: %s", err)
			}

			if err = writeReplacement(tt.args.ranges, dir, tt.args.replacement); (err != nil) != tt.wantErr {
				t.Errorf("writeReplacement() error = %v, wantErr %v", err, tt.wantErr)
			}

			for fileName, expected := range tt.expectedContent {
				var read []byte
				read, err = ioutil.ReadFile(dir + "/" + fileName)
				if err != nil {
					t.Fatalf("couldn't read the output file: %s", err)
				}

				created := string(read)
				if created != expected {
					t.Errorf("created output file %s does not match the expected one %s", created, expected)
				}
			}
		})
	}
}
