import type { PageLoad } from './$types'
import { CommitPage_CommitQuery, CommitPage_DiffQuery } from './page.gql'

const PAGE_SIZE = 20

export const load: PageLoad = async ({ parent, params }) => {
    const {
        resolvedRevision: { repo },
        graphqlClient,
    } = await parent()

    const commit = await graphqlClient
        .query({ query: CommitPage_CommitQuery, variables: { repo: repo.id, revspec: params.revspec } })
        .then(result => {
            if (result.data.node?.__typename === 'Repository') {
                return result.data.node.commit
            }
            return null
        })

    const diff =
        commit?.oid && commit?.parents[0]?.oid
            ? graphqlClient.watchQuery({
                  query: CommitPage_DiffQuery,
                  variables: {
                      repo: repo.id,
                      base: commit.parents[0].oid,
                      head: commit.oid,
                      first: PAGE_SIZE,
                      after: null,
                  },
              })
            : null

    if (diff && !graphqlClient.readQuery({ query: CommitPage_DiffQuery, variables: diff.variables })) {
        // Eagerly fetch data if it isn't in the cache already. This ensures that the data is fetched
        // as soon as possible, not only after the layout subscribes to the query.
        diff.refetch()
    }

    return {
        commit,
        diff,
    }
}
