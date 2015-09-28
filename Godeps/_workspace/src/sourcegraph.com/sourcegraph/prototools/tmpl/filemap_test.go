package tmpl

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestFilemap(t *testing.T) {
	tst := `
    <FileMap>
        <Generate>
            <Template>path/to/template.html</Template>
            <Target>path/to/target.proto</Target>
            <Output>path/to/output.html</Output>
            <Includes>
                <Include>a.tmpl</Include>
                <Include>b.tmpl</Include>
            </Includes>
            <Data>
                <Item>
                    <Key>key1</Key>
                    <Value>value1</Value>
                </Item>
                <Item>
                    <Key>key2</Key>
                    <Value>value2</Value>
                </Item>
                <Item>
                    <Key>key3</Key>
                    <Value>value3</Value>
                </Item>
            </Data>
        </Generate>
    </FileMap>
    `
	want := FileMap{
		Generate: []*FileMapGenerate{
			{
				Template: "path/to/template.html",
				Target:   "path/to/target.proto",
				Output:   "path/to/output.html",
				Include:  []string{"a.tmpl", "b.tmpl"},
				Data: []*FileMapDataItem{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
					{Key: "key3", Value: "value3"},
				},
			},
		},
	}

	// Verify the unmarshaled data is exactly equal.
	var f FileMap
	err := xml.Unmarshal([]byte(tst), &f)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f, want) {
		// Print the XML we expect, since the test failed.
		wantOut, err := xml.MarshalIndent(want, "", "    ")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("\n%s\n", string(wantOut))
		t.Fatal("not equal")
	}
}
