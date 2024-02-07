import { getGraphQLClient } from '$lib/graphql'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { TagsPage_TagsQuery } from './page.gql'

export const load: PageLoad = async ({ params }) => {
    const client = await getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)

    return {
        tags: client
            .query({
                query: TagsPage_TagsQuery,
                variables: {
                    repoName,
                    first: 20,
                    withBehindAhead: false,
                },
            })
            .then(result => {
                if (!result.data.repository) {
                    throw new Error('Expected Repository')
                }
                return result.data.repository.gitRefs
            }),
    }
}
