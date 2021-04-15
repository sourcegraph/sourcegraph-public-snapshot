import { Observable, of } from 'rxjs'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { ISearchContext } from '@sourcegraph/shared/src/graphql/schema'

import { ListSearchContextsResult } from '../graphql-operations'

export function mockFetchAutoDefinedSearchContexts(): Observable<ISearchContext[]> {
    return of([] as ISearchContext[])
}

export function mockFetchSearchContexts({
    first,
    includeAll,
    namespace,
    query,
    after,
}: {
    first: number
    query?: string
    namespace?: Scalars['ID']
    includeAll?: boolean
    after?: string
}): Observable<ListSearchContextsResult['searchContexts']> {
    const result: ListSearchContextsResult['searchContexts'] = {
        nodes: [],
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        totalCount: 0,
    }
    return of(result)
}
