package main

func search(query string) (GQLResult, error) {
	response, err := post(query, "/.api/graphql?Search")
	if err != nil {
		return "", err
	}
	return response, nil
}

const gqlSearch = `query Search($query: String!) {
	search(query: $query) {
		results {
			limitHit
			matchCount
			results {
				__typename
				... on Repository {
					name
				}
				... on FileMatch {
					resource
					limitHit
					lineMatches {
						preview
						lineNumber
						offsetAndLengths
					}
				}
				... on CommitSearchResult {
					refs {
						name
						displayName
						prefix
						repository { uri }
					}
					sourceRefs {
						name
						displayName
						prefix
						repository { uri }
					}
					messagePreview {
						value
						highlights {
							line
							character
							length
						}
					}
					diffPreview {
						value
						highlights {
							line
							character
							length
						}
					}
					commit {
						repository {
							uri
						}
						oid
						abbreviatedOID
						author {
							person {
								displayName
								avatarURL
							}
							date
						}
						message
					}
				}
			}
			alert {
				title
				proposedQueries {
					description
					query
				}
			}
			dynamicFilters {
				value
				count
			}
		}
	}
}
`
