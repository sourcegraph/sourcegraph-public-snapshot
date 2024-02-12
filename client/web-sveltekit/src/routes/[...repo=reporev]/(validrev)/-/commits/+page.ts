import { getGraphQLClient } from '$lib/graphql'
import { resolveRevision } from '$lib/repo/utils'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { CommitsPage_CommitsQuery } from './page.gql'

const PAGE_SIZE = 20

export const load: PageLoad = async ({ parent, params }) => {
    const client = await getGraphQLClient()
    const { repoName, revision = '' } = parseRepoRevision(params.repo)
    const resolvedRevision = await resolveRevision(parent, revision)

    const commitsQuery = client.watchQuery({
        query: CommitsPage_CommitsQuery,
        variables: {
            repoName,
            revision: resolvedRevision,
            first: PAGE_SIZE,
            afterCursor: null,
        },
        notifyOnNetworkStatusChange: true,
    })

    if (!client.readQuery({ query: CommitsPage_CommitsQuery, variables: commitsQuery.variables })) {
        // Eagerly fetch data if it isn't in the cache already. This ensures that the data is fetched
        // as soon as possible, not only after the layout subscribes to the query.
        commitsQuery.refetch()
    }

    return {
        commitsQuery,
    }
}
