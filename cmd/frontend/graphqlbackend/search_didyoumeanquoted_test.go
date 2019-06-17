package graphqlbackend

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func Test_didYouMeanQuotedResolver_Results(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		query string
		err   error
	}
	tests := []struct {
		name    string
		fields  fields
		want    *searchResultsResolver
		wantErr bool
	}{
		{
			name: "empty",
			fields: fields{
				query: "",
			},
			want: &searchResultsResolver{
				alert: &searchAlert{
					title: "Try quoted",
					proposedQueries: []*searchQueryDescription{
						{
							description: "query quoted entirely",
							query:       `""`,
						},
						{
							description: "query with parts quoted",
							query:       "",
						},
					},
				},
			},
		},
		{
			name: `fmt.Sprintf("`,
			fields: fields{
				query: `fmt.Sprintf("`,
				err:   errors.New(`type error at character 0: error parsing regexp: missing closing ):`),
			},
			want: &searchResultsResolver{
				alert: &searchAlert{
					title:       "Try quoted",
					description: `type error at character 0: error parsing regexp: missing closing ):`,
					proposedQueries: []*searchQueryDescription{
						{
							description: "query quoted entirely",
							query:       `"fmt.Sprintf(\""`,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &didYouMeanQuotedResolver{
				query: tt.fields.query,
				err:   tt.fields.err,
			}
			got, err := r.Results(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("didYouMeanQuotedResolver.Results() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.alert.proposedQueries, tt.want.alert.proposedQueries) {
				t.Errorf("didYouMeanQuotedResolver.Results().alert.proposedQueries =\n%s\nwant\n%s", spew.Sdump(got.alert.proposedQueries), spew.Sdump(tt.want.alert.proposedQueries))
			}
			if !reflect.DeepEqual(got.alert, tt.want.alert) {
				t.Errorf("didYouMeanQuotedResolver.Results().alert =\n%+v\nwant\n%+v", got.alert, tt.want.alert)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("didYouMeanQuotedResolver.Results() =\n%+v\nwant\n%+v", got, tt.want)
			}
		})
	}
}
