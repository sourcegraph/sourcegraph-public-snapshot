import { getGraphQLClient } from '$lib/graphql'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { CommitPage_CommitQuery, CommitPage_DiffQuery } from './page.gql'

const PAGE_SIZE = 20

export const load: PageLoad = async ({ params }) => {
    const client = await getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)

    const commit = await client
        .query({ query: CommitPage_CommitQuery, variables: { repoName, revspec: params.revspec } })
        .then(result => {
            return result.data.repository?.commit ?? null
        })

    const diff =
        commit?.oid && commit?.parents[0]?.oid
            ? client.watchQuery({
                  query: CommitPage_DiffQuery,
                  variables: {
                      repoName,
                      base: commit.parents[0].oid,
                      head: commit.oid,
                      first: PAGE_SIZE,
                      after: null,
                  },
              })
            : null

    if (diff && !client.readQuery({ query: CommitPage_DiffQuery, variables: diff.variables })) {
        // Eagerly fetch data if it isn't in the cache already. This ensures that the data is fetched
        // as soon as possible, not only after the layout subscribes to the query.
        diff.refetch()
    }

    return {
        commit,
        diff,
    }
}
