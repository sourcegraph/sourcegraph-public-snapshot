import { gql, query } from '$lib/graphql'
import type {
    PagedRepositoryContributorConnectionFields,
    PagedRepositoryContributorsResult,
    PagedRepositoryContributorsVariables,
} from '$lib/graphql-operations'

const CONTRIBUTORS_QUERY = gql`
    query PagedRepositoryContributors(
        $repo: ID!
        $first: Int
        $last: Int
        $after: String
        $before: String
        $revisionRange: String
        $afterDate: String
        $path: String
    ) {
        node(id: $repo) {
            __typename
            id
            ... on Repository {
                contributors(
                    first: $first
                    last: $last
                    before: $before
                    after: $after
                    revisionRange: $revisionRange
                    afterDate: $afterDate
                    path: $path
                ) {
                    ...PagedRepositoryContributorConnectionFields
                }
            }
        }
    }

    fragment PagedRepositoryContributorConnectionFields on RepositoryContributorConnection {
        totalCount
        pageInfo {
            hasNextPage
            hasPreviousPage
            endCursor
            startCursor
        }
        nodes {
            ...PagedRepositoryContributorNodeFields
        }
    }

    fragment PagedRepositoryContributorNodeFields on RepositoryContributor {
        __typename
        person {
            name
            displayName
            email
            avatarURL
            user {
                username
                url
                displayName
                avatarURL
            }
        }
        count
        commits(first: 1) {
            nodes {
                id
                oid
                abbreviatedOID
                canonicalURL
                subject
                author {
                    date
                }
            }
        }
    }
`

const emptyPage: Extract<PagedRepositoryContributorsResult['node'], { __typename: 'Repository' }>['contributors'] = {
    totalCount: 0,
    nodes: [] as any[],
    pageInfo: {
        hasNextPage: false,
        hasPreviousPage: false,
        endCursor: null,
        startCursor: null,
    },
}

export async function fetchContributors(
    options: PagedRepositoryContributorsVariables
): Promise<PagedRepositoryContributorConnectionFields> {
    const data = await query<PagedRepositoryContributorsResult, PagedRepositoryContributorsVariables>(
        CONTRIBUTORS_QUERY,
        options
    )

    if (data?.node?.__typename === 'Repository') {
        return data.node.contributors
    }

    return emptyPage
}
