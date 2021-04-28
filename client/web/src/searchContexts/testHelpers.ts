import { Observable, of } from 'rxjs'

import { Scalars, SearchContextsNamespaceFilterType } from '@sourcegraph/shared/src/graphql-operations'
import { ISearchContext } from '@sourcegraph/shared/src/graphql/schema'

import { ListSearchContextsResult } from '../graphql-operations'

export function mockFetchAutoDefinedSearchContexts(): Observable<ISearchContext[]> {
    return of([] as ISearchContext[])
}

export function mockFetchSearchContexts({
    first,
    filterType,
    namespace,
    query,
    after,
}: {
    first: number
    query?: string
    namespace?: Scalars['ID']
    filterType?: SearchContextsNamespaceFilterType
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
