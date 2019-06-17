package graphqlbackend

import (
	"context"
	"reflect"
	"testing"
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("didYouMeanQuotedResolver.Results() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
