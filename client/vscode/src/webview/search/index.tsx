import React, { useMemo } from 'react'

import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { WebviewPageProps } from '..'
import { SearchResult, SearchVariables } from '../../graphql-operations'

import styles from './index.module.scss'
import { SearchResults } from './SearchResults'

interface SearchPageProps extends WebviewPageProps {}

const searchQuery = gql`
    query Search($query: String!) {
        search(version: V2, query: $query) {
            results {
                matchCount
                results {
                    ... on FileMatch {
                        lineMatches {
                            preview
                        }
                    }
                    ... on CommitSearchResult {
                        matches {
                            body {
                                text
                            }
                        }
                    }
                    ... on Repository {
                        name
                        description
                    }
                }
            }
        }
    }
`

export const SearchPage: React.FC<SearchPageProps> = ({ platformContext }) => {
    const searchResults = useObservable(
        useMemo(
            () =>
                platformContext.requestGraphQL<SearchResult, SearchVariables>({
                    request: searchQuery,
                    variables: { query: 'repo:github.com/sourcegraph comlink' },
                    mightContainPrivateInfo: false,
                }),
            [platformContext]
        )
    )

    console.log({ searchResults })

    return (
        <div className={styles.title}>
            <h1>SearchWebview</h1>
            <input type="text" placeholder="Search" />

            <SearchResults />
        </div>
    )
}
