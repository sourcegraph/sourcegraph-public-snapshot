import { getPaginationParams } from '$lib/Paginator'

import type { PageLoad } from './$types'
import { CommitsQuery } from './page.gql'

const pageSize = 20

export const load: PageLoad = async ({ parent, url }) => {
    const { resolvedRevision, graphqlClient } = await parent()
    const { first, after } = getPaginationParams(url.searchParams, pageSize)

    return {
        deferred: {
            commits: graphqlClient
                .query({
                    query: CommitsQuery,
                    variables: {
                        repo: resolvedRevision.repo.id,
                        revspec: resolvedRevision.commitID,
                        first,
                        afterCursor: after,
                    },
                })
                .then(result => {
                    if (result.data.node?.__typename !== 'Repository') {
                        throw new Error('Unable to find repository')
                    }
                    if (!result.data.node.commit) {
                        throw new Error('Unable to find commit')
                    }
                    return result.data.node.commit.ancestors
                }),
        },
    }
}
