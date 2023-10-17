import { subDays } from 'date-fns'
import { type Observable, of } from 'rxjs'

import type { AuthenticatedUser } from '../../auth'
import type { Maybe, Scalars } from '../../graphql-operations'

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
    viewerHasAsDefault: boolean
    viewerHasStarred: boolean
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

export function mockFetchSearchContexts(): Observable<ListSearchContexts> {
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
                viewerHasAsDefault: false,
                viewerHasStarred: false,
                description: 'All code on Sourcegraph',
                repositories: [],
                query: '',
                updatedAt: subDays(new Date(), 1).toISOString(),
            },
            {
                __typename: 'SearchContext',
                id: '1',
                spec: 'test',
                name: 'test',
                namespace: null,
                public: true,
                autoDefined: false,
                viewerCanManage: true,
                viewerHasAsDefault: false,
                viewerHasStarred: true,
                description: 'Test context',
                repositories: [],
                query: '',
                updatedAt: subDays(new Date(), 1).toISOString(),
            },
            {
                __typename: 'SearchContext',
                id: '2',
                spec: '@user/usertest',
                name: 'usertest',
                namespace: {
                    __typename: 'User',
                    id: '1',
                    namespaceName: 'user',
                },
                public: false,
                autoDefined: false,
                viewerCanManage: true,
                viewerHasAsDefault: true,
                viewerHasStarred: false,
                description: '',
                repositories: [],
                query: '',
                updatedAt: subDays(new Date(), 2).toISOString(),
            },
        ],
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        totalCount: 3,
    }
    return of(result)
}

export function mockGetUserSearchContextNamespaces(): Maybe<Scalars['ID']>[] {
    return []
}

export const mockAuthenticatedUser = {
    __typename: 'User',
    id: '1',
    username: 'user',
    organizations: {
        __typename: 'OrgConnection',
        nodes: [],
    },
    permissions: {
        __typename: 'PermissionConnection',
        nodes: [],
    },
} as unknown as AuthenticatedUser
