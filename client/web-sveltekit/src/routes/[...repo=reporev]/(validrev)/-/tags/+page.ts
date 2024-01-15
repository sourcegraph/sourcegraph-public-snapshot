import type { PageLoad } from './$types'
import { GitTagsQuery } from './page.gql'

export const load: PageLoad = async ({ parent }) => {
    const { resolvedRevision, graphqlClient } = await parent()

    return {
        deferred: {
            tags: graphqlClient
                .query({
                    query: GitTagsQuery,
                    variables: {
                        repo: resolvedRevision.repo.id,
                        first: 20,
                        withBehindAhead: false,
                    },
                })
                .then(result => {
                    if (result.data.node?.__typename !== 'Repository') {
                        throw new Error('Expected Repository')
                    }
                    return result.data.node.gitRefs
                }),
        },
    }
}
