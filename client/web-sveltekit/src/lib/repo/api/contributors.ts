import type {
    PagedRepositoryContributorConnectionFields,
    PagedRepositoryContributorsResult,
    PagedRepositoryContributorsVariables,
} from '$lib/graphql-operations'
import { getDocumentNode, gql, type GraphQLClient } from '$lib/http-client'

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
            ... on Repository {
                __typename
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
                oid
                abbreviatedOID
                url
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
    client: GraphQLClient,
    options: PagedRepositoryContributorsVariables
): Promise<PagedRepositoryContributorConnectionFields> {
    const response = await client.query<PagedRepositoryContributorsResult, PagedRepositoryContributorsVariables>({
        query: getDocumentNode(CONTRIBUTORS_QUERY),
        variables: options,
    })

    if (response.data?.node?.__typename === 'Repository') {
        return response.data.node.contributors
    }

    return emptyPage
}
