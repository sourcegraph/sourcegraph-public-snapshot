package makex

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		data         string
		wantMakefile *Makefile
		wantErr      error
	}{
		"empty ": {data: ``, wantMakefile: &Makefile{}},
		"rule with 1 target, 1 prereq": {
			data:         `x:y`,
			wantMakefile: &Makefile{Rules: []Rule{&BasicRule{"x", []string{"y"}, nil}}},
		},
		"rule with multiple targets": {
			data:    `x0 x1:y`,
			wantErr: errMultipleTargetsUnsupported(0),
		},
		"rule with multiple prereqs": {
			data:         `x : y0 y1`,
			wantMakefile: &Makefile{Rules: []Rule{&BasicRule{"x", []string{"y0", "y1"}, nil}}},
		},
		"rule with duplicate prereqs": {
			data:         `x : y0 y1 y0 y1 y1`,
			wantMakefile: &Makefile{Rules: []Rule{&BasicRule{"x", []string{"y0", "y1"}, nil}}},
		},
		"multiple rules": {
			data: `
x0:y0
x1:y1`,
			wantMakefile: &Makefile{Rules: []Rule{&BasicRule{"x0", []string{"y0"}, nil}, &BasicRule{"x1", []string{"y1"}, nil}}},
		},
		"rule with recipes": {
			data: `
x:y
	c0
	c1`,
			wantMakefile: &Makefile{Rules: []Rule{&BasicRule{"x", []string{"y"}, []string{"c0", "c1"}}}},
		},
		"multiple rules with recipes": {
			data: `
x0:y0
	c0
a = 3
x1:y1
	c1`,
			wantMakefile: &Makefile{Rules: []Rule{
				&BasicRule{"x0", []string{"y0"}, []string{"c0"}},
				&BasicRule{"x1", []string{"y1"}, []string{"c1"}},
			}},
		},
		"recipe with $@ (target) var": {
			data: `
x:
	echo $@`,
			wantMakefile: &Makefile{Rules: []Rule{&BasicRule{"x", []string{}, []string{"echo x"}}}},
		},
		"recipe with $^ (prereqs) var": {
			data: `
x: a b
	echo $^`,
			wantMakefile: &Makefile{Rules: []Rule{&BasicRule{"x", []string{"a", "b"}, []string{"echo a b"}}}},
		},
	}
	for label, test := range tests {
		mf, err := Parse([]byte(test.data))
		if !reflect.DeepEqual(err, test.wantErr) {
			if test.wantErr == nil {
				t.Errorf("%s: Parse: error: %s", label, err)
				continue
			} else {
				t.Errorf("%s: Parse: error: got %q, want %q", label, err, test.wantErr)
				continue
			}
		}
		if !reflect.DeepEqual(mf, test.wantMakefile) {
			t.Errorf("%s: bad parsed Makefile\n=========== got Makefile\n%s\n\n=========== want Makefile\n%s", label, marshalStr(t, mf), marshalStr(t, test.wantMakefile))
		}
	}
}

func marshalStr(t *testing.T, mf *Makefile) string {
	data, err := Marshal(mf)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(data))
}
