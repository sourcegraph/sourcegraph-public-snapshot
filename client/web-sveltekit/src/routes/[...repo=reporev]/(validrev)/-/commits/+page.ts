import type { PageLoad } from './$types'
import { CommitsPage_CommitsQuery } from './page.gql'

const PAGE_SIZE = 20

export const load: PageLoad = async ({ parent }) => {
    const { resolvedRevision, graphqlClient } = await parent()

    const commitsQuery = graphqlClient.watchQuery({
        query: CommitsPage_CommitsQuery,
        variables: {
            repo: resolvedRevision.repo.id,
            revspec: resolvedRevision.commitID,
            first: PAGE_SIZE,
            afterCursor: null,
        },
        notifyOnNetworkStatusChange: true,
    })

    if (!graphqlClient.readQuery({ query: CommitsPage_CommitsQuery, variables: commitsQuery.variables })) {
        // Eagerly fetch data if it isn't in the cache already. This ensures that the data is fetched
        // as soon as possible, not only after the layout subscribes to the query.
        commitsQuery.refetch()
    }

    return {
        commitsQuery,
    }
}
