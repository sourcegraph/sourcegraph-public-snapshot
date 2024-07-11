import { getGraphQLClient, mapOrThrow } from '$lib/graphql'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { BranchesPage_OverviewQuery } from './page.gql'

export const load: PageLoad = ({ params }) => {
    const client = getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)

    return {
        overview: client
            .query(BranchesPage_OverviewQuery, {
                first: 20,
                repoName,
                withBehindAhead: true,
            })
            .toPromise()
            .then(
                mapOrThrow(result => {
                    if (!result.data?.repository) {
                        // This page will never render when the repository is not found.
                        // The (validrev) data loader will render an error page instead.
                        throw new Error('Unable to load repository data.')
                    }
                    return result.data.repository
                })
            ),
    }
}
