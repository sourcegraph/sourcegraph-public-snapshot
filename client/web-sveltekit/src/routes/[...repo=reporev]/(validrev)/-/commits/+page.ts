import { getPaginationParams } from '$lib/Paginator'
import {CommitsQueryDocument} from './page.gql.ts'

import type { PageLoad } from './$types'

const pageSize = 20

export const load: PageLoad = async ({ parent, url }) => {
    const { resolvedRevision, graphqlClient } = await parent()
    const { first, after } = getPaginationParams(url.searchParams, pageSize)

    return {
        deferred: {
            commits: graphqlClient.query({
                query: CommitsQueryDocument,
                variables: {
                    repo: resolvedRevision.repo.id,
                    revspec: resolvedRevision.commitID,
                    first,
                    afterCursor: after
                },
            }).then(result => {
                if (result.data.node?.__typename !== 'Repository') {
                    throw new Error('Unable to find repository')
                }
                if (!result.data.node.commit) {
                    throw new Error('Unable to find commit')
                }
                return result.data.node.commit.ancestors
            })
        },
    }
}
