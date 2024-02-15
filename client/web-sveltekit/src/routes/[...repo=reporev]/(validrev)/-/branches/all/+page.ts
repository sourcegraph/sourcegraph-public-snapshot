import { getGraphQLClient } from '$lib/graphql'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { AllBranchesPage_BranchesQuery } from './page.gql'

export const load: PageLoad = async ({ params }) => {
    const client = await getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)

    return {
        branches: client
            .query({
                query: AllBranchesPage_BranchesQuery,
                variables: {
                    repoName,
                    first: 20,
                    withBehindAhead: true,
                },
            })
            .then(result => {
                if (!result.data.repository) {
                    // This page will never render when the repository is not found.
                    // The (validrev) data loader will render an error page instead.
                    // Still, this error will show up as an unhandled promise rejection
                    // in the console. We should find a better way to handle this.
                    throw new Error('Expected Repository')
                }
                return result.data.repository.branches
            }),
    }
}
