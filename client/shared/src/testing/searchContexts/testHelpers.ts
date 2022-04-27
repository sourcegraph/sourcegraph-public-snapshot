import { subDays } from 'date-fns'
import { range } from 'lodash'
import { Observable, of } from 'rxjs'

import { Maybe, Scalars } from '../../graphql-operations'
import { ISearchContext } from '../../schema'

interface SearchContextFields {
    __typename: 'SearchContext'
    id: string
    name: string
    spec: string
    description: string
    public: boolean
    autoDefined: boolean
    updatedAt: string
    viewerCanManage: boolean
    namespace: Maybe<
        | { __typename: 'User'; id: string; namespaceName: string }
        | { __typename: 'Org'; id: string; namespaceName: string }
    >
    query: string
    repositories: {
        __typename: 'SearchContextRepositoryRevisions'
        revisions: string[]
        repository: { name: string }
    }[]
}

interface ListSearchContexts {
    totalCount: number
    nodes: SearchContextFields[]
    pageInfo: { hasNextPage: boolean; endCursor: Maybe<string> }
}

export function mockFetchAutoDefinedSearchContexts(numberContexts = 0): () => Observable<ISearchContext[]> {
    return () =>
        of(
            range(0, numberContexts).map(index => ({
                __typename: 'SearchContext',
                id: index.toString(),
                spec: `auto-defined-${index}`,
                name: `auto-defined-${index}`,
                namespace: null,
                public: true,
                autoDefined: true,
                viewerCanManage: false,
                description: 'Repositories on Sourcegraph',
                repositories: [],
                query: '',
                updatedAt: subDays(new Date(), 1).toISOString(),
            })) as ISearchContext[]
        )
}

export function mockFetchSearchContexts({
    first,
    query,
    after,
}: {
    first: number
    query?: string
    after?: string
}): Observable<ListSearchContexts> {
    const result: ListSearchContexts = {
        nodes: [],
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        totalCount: 0,
    }
    return of(result)
}

export function mockGetUserSearchContextNamespaces(): Maybe<Scalars['ID']>[] {
    return []
}
