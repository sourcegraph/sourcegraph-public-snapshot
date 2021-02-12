package graphqlbackend

import "testing"

func TestEstimateQueryCost(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query string
		want  int
	}{
		{
			name: "Canonical example",
			query: `query {
  viewer {
    login
    repositories(first: 100) {
      edges {
        node {
          id
          issues(first: 50) {
            edges {
              node {
                id
                labels(first: 60) {
                  edges {
                    node {
                      id
                      name
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`,
			want: 51,
		},
		{
			name: "simple query",
			query: `
query {
  viewer {
    repositories(first: 50) {
      edges {
        repository:node {
          name
          issues(first: 10) {
            totalCount
            edges {
              node {
                title
                bodyHTML
              }
            }
          }
        }
      }
    }
  }
}
`,
			want: 1,
		},
		{
			name: "complex query",
			query: `query {
  viewer {
    repositories(first: 50) {
      edges {
        repository:node {
          name

          pullRequests(first: 20) {
            edges {
              pullRequest:node {
                title

                comments(first: 10) {
                  edges {
                    comment:node {
                      bodyHTML
                    }
                  }
                }
              }
            }
          }

          issues(first: 20) {
            totalCount
            edges {
              issue:node {
                title
                bodyHTML

                comments(first: 10) {
                  edges {
                    comment:node {
                      bodyHTML
                    }
                  }
                }
              }
            }
          }
        }
      }
    }

    followers(first: 10) {
      edges {
        follower:node {
          login
        }
      }
    }
  }
}`,
			want: 21,
		},
		{
			name: "Multiple top level queries",
			query: `query {
  thing
}
query{
  thing
}
`,
			want: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			have, err := estimateQueryCost(tc.query)
			if err != nil {
				t.Fatal(err)
			}
			if have != tc.want {
				t.Fatalf("have %d, want %d", have, tc.want)
			}
		})
	}
}
