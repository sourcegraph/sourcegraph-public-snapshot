package api

import (
	"encoding/json"
	"testing"
)

func TestGraphQLError_Code(t *testing.T) {
	for name, tc := range map[string]struct {
		in      string
		want    string
		wantErr bool
	}{
		"invalid code": {
			in: `{
				"errors": [
					{
						"message": "The feature \"campaigns\" is not activated because it requires a valid Sourcegraph license. Purchase a Sourcegraph subscription to activate this feature.",
						"path": [
							"createBatchSpec"
						],
						"extensions": {
							"code": 42
						}
					}
				],
				"data": null
			}`,
			wantErr: true,
		},
		"invalid extensions": {
			in: `{
				"errors": [
					{
						"message": "The feature \"campaigns\" is not activated because it requires a valid Sourcegraph license. Purchase a Sourcegraph subscription to activate this feature.",
						"path": [
							"createBatchSpec"
						],
						"extensions": 42
					}
				],
				"data": null
			}`,
			wantErr: true,
		},
		"no code": {
			in: `{
				"errors": [
					{
						"message": "The feature \"campaigns\" is not activated because it requires a valid Sourcegraph license. Purchase a Sourcegraph subscription to activate this feature.",
						"path": [
							"createBatchSpec"
						],
						"extensions": {}
					}
				],
				"data": null
			}`,
			want: "",
		},
		"no extensions": {
			in: `{
				"errors": [
					{
						"message": "The feature \"campaigns\" is not activated because it requires a valid Sourcegraph license. Purchase a Sourcegraph subscription to activate this feature.",
						"path": [
							"createBatchSpec"
						]
					}
				],
				"data": null
			}`,
			want: "",
		},
		"valid code": {
			in: `{
				"errors": [
					{
						"message": "The feature \"campaigns\" is not activated because it requires a valid Sourcegraph license. Purchase a Sourcegraph subscription to activate this feature.",
						"path": [
							"createBatchSpec"
						],
						"extensions": {
							"code": "ErrBatchChangesUnlicensed"
						}
					}
				],
				"data": null
			}`,
			want: "ErrBatchChangesUnlicensed",
		},
	} {
		t.Run(name, func(t *testing.T) {
			var result rawResult
			if err := json.Unmarshal([]byte(tc.in), &result); err != nil {
				t.Fatal(err)
			}
			if ne := len(result.Errors); ne != 1 {
				t.Fatalf("unexpected number of GraphQL errors (this test can only handle one!): %d", ne)
			}

			ge := &GraphQlError{result.Errors[0]}
			have, err := ge.Code()
			if tc.wantErr {
				if err == nil {
					t.Errorf("unexpected nil error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %+v", err)
				}
				if have != tc.want {
					t.Errorf("unexpected code: have=%q want=%q", have, tc.want)
				}
			}
		})
	}

}
