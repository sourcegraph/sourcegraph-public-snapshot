import type { PageLoad } from './$types'
import { GitBranchesQuery } from './page.gql'

export const load: PageLoad = async ({ parent }) => {
    const { resolvedRevision, graphqlClient } = await parent()
    return {
        branches: graphqlClient
            .query({
                query: GitBranchesQuery,
                variables: {
                    repo: resolvedRevision.repo.id,
                    first: 20,
                    withBehindAhead: true,
                },
            })
            .then(result => {
                if (result.data.node?.__typename !== 'Repository') {
                    throw new Error('Expected Repository')
                }
                return result.data.node.branches
            }),
    }
}
