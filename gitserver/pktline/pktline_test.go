package pktline

import (
	"testing"
)

func TestParseLength(t *testing.T) {
	tests := []struct {
		data        string
		length      int
		expectError bool
	}{
		{"0000", 0, false},
		{"0005a", 5, false},
		{"0228aaaaaaaaaaaaaa", 552, false},
		{"", 0, true},
		{"00\n00", 0, true},
		{"00xx", 0, true},
		{"xx00", 0, true},
		{"xxxx", 0, true},
	}
	for _, tt := range tests {
		length, err := parseLength([]byte(tt.data))
		switch {
		case tt.expectError && err == nil:
			t.Errorf("Expected error for %q", tt.data)
		case !tt.expectError && err != nil:
			t.Errorf("Error not expected for %q: %s", tt.data, err)
		case tt.length != length:
			t.Errorf("Expected length of %d for %q, got %d", tt.length, tt.data, length)
		}
	}
}

func TestScan(t *testing.T) {
	tests := []struct {
		data        string
		atEOF       bool
		advance     int
		token       string
		expectError bool
	}{
		// Flush.
		{"0000", false, 4, "0000", false},
		{"0000", true, 4, "0000", false},
		{"0000xxxx", false, 4, "0000", false},

		// Need more.
		{"0076ca82a6dff817ec66f4", false, 0, "", false},
		{"0076ca82a6dff817ec66f44342007202690a93763949 15027957951b64cf874c3557a0f3547bd83b3ff6 refs/heads/master report-status\n",
			false, 118, "0076ca82a6dff817ec66f44342007202690a93763949 15027957951b64cf874c3557a0f3547bd83b3ff6 refs/heads/master report-status\n", false},
		{"0076ca82a6dff817ec66f44342007202690a93763949 15027957951b64cf874c3557a0f3547bd83b3ff6 refs/heads/master report-status\n0000",
			false, 118, "0076ca82a6dff817ec66f44342007202690a93763949 15027957951b64cf874c3557a0f3547bd83b3ff6 refs/heads/master report-status\n", false},

		// At EOF.
		{"", true, 0, "", false},
		{"000", true, 0, "", true},
		{"0076ca82a6dff817ec66f44342007202690a93763949 15027957951b64cf874c3557a0f3547bd83b3ff6 refs/heads/master report-status\n",
			true, 118, "0076ca82a6dff817ec66f44342007202690a93763949 15027957951b64cf874c3557a0f3547bd83b3ff6 refs/heads/master report-status\n", false},
		// Invalid.
		{"00xxca82a6dff817ec66f44342007202690a93763949 15027957951b64cf874c3557a0f3547bd83b3ff6 refs/heads/master report-status\n",
			true, 0, "", true},
	}
	for _, tt := range tests {
		advance, token, err := SplitFunc([]byte(tt.data), tt.atEOF)
		switch {
		case tt.expectError && err == nil:
			t.Errorf("Expected error for %q", tt.data)
		case !tt.expectError && err != nil:
			t.Errorf("Error not expected for %q: %s", tt.data, err)
		case tt.advance != advance:
			t.Errorf("Expected advance of %d for %q, got %d", tt.advance, tt.data, advance)
		case tt.token != string(token):
			t.Errorf("Expected token of %q for %q, got %q", tt.token, tt.data, string(token))
		}
	}
}

func TestIsComment(t *testing.T) {
	tests := []struct {
		data      string
		isComment bool
	}{
		{"0005#", true},
		{"00xx#safsdf", true},
		{"0005a", false},
		{"xx", false},
	}
	for _, tt := range tests {
		isComment := IsComment([]byte(tt.data))
		if tt.isComment && !isComment {
			t.Errorf("Expected %q to be a comment", tt.data)
		} else if !tt.isComment && isComment {
			t.Errorf("Did not expect %q to be a comment", tt.data)
		}
	}
}

func TestIsFlush(t *testing.T) {
	tests := []struct {
		data    string
		isFlush bool
	}{
		{"0000", true},
		{"0000\n", false},
		{"000xx", false},
		{"xx", false},
	}
	for _, tt := range tests {
		isFlush := IsFlush([]byte(tt.data))
		if tt.isFlush && !isFlush {
			t.Errorf("Expected %q to be a flush", tt.data)
		} else if !tt.isFlush && isFlush {
			t.Errorf("Did not expect %q to be a flush", tt.data)
		}
	}
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		data      string
		prefix    string
		hasPrefix bool
	}{
		{"00xxtest", "test", true},
		{"test", "test", false},
	}
	for _, tt := range tests {
		hasPrefix := HasPrefix([]byte(tt.data), []byte(tt.prefix))
		if tt.hasPrefix && !hasPrefix {
			t.Errorf("Expected %q to have the prefix %q", tt.data, tt.prefix)
		} else if !tt.hasPrefix && hasPrefix {
			t.Errorf("Did not expect %q to have the prefix %q", tt.data, tt.prefix)
		}
	}
}

func TestCommandType(t *testing.T) {
	tests := []struct {
		data                         string
		isCreate, isDelete, isUpdate bool
	}{
		// Invalid.
		{data: "xxxx0000000000000000000000000000000000000000 xx /ref/xxxx"},
		{data: "xxxxxx 0000000000000000000000000000000000000000 /ref/xxxx"},
		{isCreate: true, data: "xxxx0000000000000000000000000000000000000000 xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx /ref/xxxx"},
		{isCreate: true, data: "xxxx0000000000000000000000000000000000000000 xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx /ref/xxxx\n"},
		{isDelete: true, data: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx 0000000000000000000000000000000000000000 /ref/xxxx"},
		{isDelete: true, data: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx 0000000000000000000000000000000000000000 /ref/xxxx\n"},
		{isUpdate: true, data: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy /ref/xxxx"},
		{isUpdate: true, data: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy /ref/xxxx\n"},
	}
	for _, tt := range tests {
		isCreate := IsCreate([]byte(tt.data))
		if tt.isCreate && !isCreate {
			t.Errorf("Expected %q to be a create command", tt.data)
		} else if !tt.isCreate && isCreate {
			t.Errorf("Did not expect %q to be a create command", tt.data)
		}
		isDelete := IsDelete([]byte(tt.data))
		if tt.isDelete && !isDelete {
			t.Errorf("Expected %q to be a delete command", tt.data)
		} else if !tt.isDelete && isDelete {
			t.Errorf("Did not expect %q to be a delete command", tt.data)
		}
		isUpdate := IsUpdate([]byte(tt.data))
		if tt.isUpdate && !isUpdate {
			t.Errorf("Expected %q to be a update command", tt.data)
		} else if !tt.isUpdate && isUpdate {
			t.Errorf("Did not expect %q to be a update command", tt.data)
		}
	}
}
