import { subDays } from 'date-fns'
import { Observable, of } from 'rxjs'

import { Maybe, Scalars } from '../../graphql-operations'

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
    viewerHasAsDefault: boolean
    viewerHasStarred: boolean
}

interface ListSearchContexts {
    totalCount: number
    nodes: SearchContextFields[]
    pageInfo: { hasNextPage: boolean; endCursor: Maybe<string> }
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
        nodes: [
            {
                __typename: 'SearchContext',
                id: '0',
                spec: 'global',
                name: 'global',
                namespace: null,
                public: true,
                autoDefined: true,
                viewerCanManage: false,
                description: 'All code on Sourcegraph',
                repositories: [],
                query: '',
                updatedAt: subDays(new Date(), 1).toISOString(),
                viewerHasAsDefault: false,
                viewerHasStarred: false,
            },
        ],
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
