import type { PageLoad } from './$types'
import { GitBranchesOverviewQuery } from './page.gql'

export const load: PageLoad = async ({ parent }) => {
    const { resolvedRevision, graphqlClient } = await parent()
    return {
        overview: graphqlClient
            .query({
                query: GitBranchesOverviewQuery,
                variables: {
                    first: 20,
                    repo: resolvedRevision.repo.id,
                    withBehindAhead: true,
                },
            })
            .then(result => {
                if (result.data.node?.__typename !== 'Repository') {
                    throw new Error('Expected Repository')
                }
                return result.data.node
            }),
    }
}
