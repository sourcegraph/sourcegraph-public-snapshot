import { getGraphQLClient, infinityQuery } from '$lib/graphql'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { TagsPage_TagsQuery } from './page.gql'

const PAGE_SIZE = 50

export const load: PageLoad = ({ params, url }) => {
    const client = getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)
    const query = url.searchParams.get('query') ?? ''

    return {
        query,
        tagsQuery: infinityQuery({
            client,
            query: TagsPage_TagsQuery,
            variables: {
                repoName,
                first: PAGE_SIZE,
                withBehindAhead: false,
                query,
            },
            nextVariables: previousResult => {
                if (previousResult?.data?.repository?.gitRefs?.pageInfo?.hasNextPage) {
                    return {
                        first: previousResult.data.repository.gitRefs.nodes.length + PAGE_SIZE,
                    }
                }
                return undefined
            },
            combine: (_previousResult, nextResult) => {
                return nextResult
            },
        }),
    }
}
