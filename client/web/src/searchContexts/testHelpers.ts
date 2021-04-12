import { Observable, of } from 'rxjs'

import { ISearchContext } from '../../../shared/src/graphql/schema'
import { ListSearchContextsResult } from '../graphql-operations'

export function mockFetchAutoDefinedSearchContexts(): Observable<ISearchContext[]> {
    return of([] as ISearchContext[])
}

export function mockFetchSearchContexts(
    first: number,
    query?: string,
    after?: string
): Observable<ListSearchContextsResult['searchContexts']> {
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
