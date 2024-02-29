import { getGraphQLClient, mapOrThrow } from '$lib/graphql'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { AllBranchesPage_BranchesQuery } from './page.gql'

export const load: PageLoad = ({ params }) => {
    const client = getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)

    return {
        branches: client
            .query(AllBranchesPage_BranchesQuery, {
                repoName,
                first: 20,
                withBehindAhead: true,
            })
            .then(
                mapOrThrow(result => {
                    if (!result.data?.repository) {
                        // This page will never render when the repository is not found.
                        // The (validrev) data loader will render an error page instead.
                        throw new Error('Expected Repository')
                    }
                    return result.data.repository.branches
                })
            ),
    }
}
