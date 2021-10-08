import React, { useMemo, useState } from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql, dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { PlatformContext, PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

export interface SearchResultsProps extends PlatformContextProps<'requestGraphQL'> {
    query: string
}

// TODO, only small selection of fields for testing
export const SEARCH_QUERY = gql`
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

const search = ({
    query,
    requestGraphQL,
}: { query: string } & Pick<PlatformContext, 'requestGraphQL'>): Observable<any> => // TODO union of error and result type
    requestGraphQL<GQL.IQuery>({
        request: SEARCH_QUERY,
        variables: { query },
        mightContainPrivateInfo: false,
    }).pipe(
        map(dataOrThrowErrors),
        map(({ search }) => {
            if (!search) {
                throw new Error('TODO error')
            }

            return search
        })
        // TODO: catchError
    )

export const SearchResults: React.FC<SearchResultsProps> = ({ query, platformContext }) => {
    const searchResults = useObservable(
        useMemo(() => search({ query, requestGraphQL: platformContext.requestGraphQL }), [
            platformContext.requestGraphQL,
            query,
        ])
    )
    console.log({ searchResults })

    if (!searchResults) {
        return <div>Loading...</div>
    }

    return <pre>{JSON.stringify({ searchResults }, null, 2)}</pre>
}
