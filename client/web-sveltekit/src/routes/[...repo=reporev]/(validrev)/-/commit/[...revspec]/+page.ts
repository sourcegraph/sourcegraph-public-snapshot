import type { PageLoad } from './$types'
import { CommitQuery, DiffQuery } from './page.gql'

export const load: PageLoad = async ({ parent, params }) => {
    const {
        resolvedRevision: { repo },
        graphqlClient,
    } = await parent()
    const commit = graphqlClient
        .query({ query: CommitQuery, variables: { repo: repo.id, revspec: params.revspec } })
        .then(result => {
            if (result.data.node?.__typename === 'Repository') {
                return result.data.node.commit
            }
            return null
        })

    return {
        deferred: {
            commit,
            // TODO: Support pagination
            diff: commit
                .then(commit => {
                    if (!commit?.oid || !commit.parents[0]?.oid) {
                        return null
                    }
                    return graphqlClient.query({
                        query: DiffQuery,
                        variables: {
                            repo: repo.id,
                            base: commit.parents[0].oid,
                            head: commit.oid,
                            paths: [],
                            first: null,
                            after: null,
                        },
                    })
                })
                .then(result => {
                    if (result?.data.node?.__typename === 'Repository') {
                        return result.data.node.comparison.fileDiffs
                    }
                    return null
                }),
        },
    }
}
