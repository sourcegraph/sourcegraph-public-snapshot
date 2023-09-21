package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestParseMount(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        *MountOptions
		expectedErr error
	}{
		{
			name:  "Valid bind mount",
			input: "type=bind,source=/foo,target=/bar",
			want: &MountOptions{
				Type:   MountTypeBind,
				Source: "/foo",
				Target: "/bar",
			},
		},
		{
			name:  "Valid readonly bind mount",
			input: "type=bind,source=/foo,target=/bar,readonly",
			want: &MountOptions{
				Type:     MountTypeBind,
				Source:   "/foo",
				Target:   "/bar",
				ReadOnly: true,
			},
		},
		{
			name:  "Valid volume mount",
			input: "type=volume,source=foo,target=/bar",
			want: &MountOptions{
				Type:   MountTypeVolume,
				Source: "foo",
				Target: "/bar",
			},
		},
		{
			name:  "Valid tmpfs mount",
			input: "type=tmpfs,target=/bar",
			want: &MountOptions{
				Type:   MountTypeTmpfs,
				Target: "/bar",
			},
		},
		{
			name:        "Invalid field",
			input:       "type=bind,source=/foo,target=/bar,baz",
			expectedErr: errors.New("invalid field 'baz' must be a key=value pair"),
		},
		{
			name:        "Invalid value for readonly field",
			input:       "type=bind,source=/foo,target=/bar,readonly=baz",
			expectedErr: errors.New("invalid value for readonly: baz"),
		},
		{
			name:        "Invalid value for bind-nonrecursive field",
			input:       "type=bind,source=/foo,target=/bar,bind-nonrecursive=baz",
			expectedErr: errors.New("invalid value for bind-nonrecursive: baz"),
		},
		{
			name:        "Invalid value for volume-nocopy field",
			input:       "type=bind,source=/foo,target=/bar,volume-nocopy=baz",
			expectedErr: errors.New("invalid value for volume-nocopy: baz"),
		},
		{
			name:        "Invalid value for tmpfs-size field",
			input:       "type=tmpfs,target=/bar,tmpfs-size=baz",
			expectedErr: errors.New("invalid value for tmpfs-size: baz"),
		},
		{
			name:        "Invalid value for tmpfs-mode field",
			input:       "type=tmpfs,target=/bar,tmpfs-mode=baz",
			expectedErr: errors.New("invalid value for tmpfs-mode: baz"),
		},
		{
			name:        "Invalid key",
			input:       "foo=baz",
			expectedErr: errors.New("unexpected key 'foo' in 'foo=baz'"),
		},
		{
			name:        "Missing target",
			input:       "source=/foo",
			expectedErr: errors.New("target is required"),
		},
		{
			name:        "Cannot mix volume options with other mount types",
			input:       "type=bind,source=/foo,destination=/bar,volume-nocopy",
			expectedErr: errors.New("cannot mix 'volume-*' options with mount type 'bind'"),
		},
		{
			name:        "Cannot mix bind options with other mount types",
			input:       "type=volume,source=foo,destination=/bar,bind-nonrecursive",
			expectedErr: errors.New("cannot mix 'bind-*' options with mount type 'volume'"),
		},
		{
			name:        "Cannot mix tmpfs options with other mount types",
			input:       "type=volume,source=/foo,destination=/bar,tmpfs-size=42",
			expectedErr: errors.New("cannot mix 'tmpfs-*' options with mount type 'volume'"),
		},
		{
			name:  "Cannot use ssh volume opt because it contains a colon",
			input: "type=volume,source=sshvolume,target=/app,volume-opt=sshcmd=test@node2:/home/test,volume-opt=password=testpassword",
			want: &MountOptions{
				Type:   MountTypeVolume,
				Source: "sshvolume",
				Target: "/app",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mount, err := ParseMount(test.input)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, test.want, mount)
		})
	}
}
