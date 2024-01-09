import { fetchTreeEntries } from '$lib/repo/api/tree'
import { findReadme } from '$lib/repo/tree'

import type { PageLoad } from './$types'
import { TreePageCommitInfoQuery, TreePageReadmeQuery } from './page.gql'

export const load: PageLoad = async ({ params, parent }) => {
    const { resolvedRevision, graphqlClient } = await parent()

    const treeEntries = fetchTreeEntries({
        repoID: resolvedRevision.repo.id,
        commitID: resolvedRevision.commitID,
        filePath: params.path,
        first: null,
    }).then(
        commit => commit.tree,
        () => null
    )

    return {
        filePath: params.path,
        deferred: {
            treeEntries,
            commitInfo: graphqlClient
                .query({
                    query: TreePageCommitInfoQuery,
                    variables: {
                        repoID: resolvedRevision.repo.id,
                        commitID: resolvedRevision.commitID,
                        filePath: params.path,
                        first: null,
                    },
                })
                .then(result => {
                    if (result.data.node?.__typename !== 'Repository') {
                        throw new Error('Unable to load repository')
                    }
                    return result.data.node.commit?.tree ?? null
                }),
            readme: treeEntries.then(result => {
                if (!result) {
                    return null
                }
                const readme = findReadme(result.entries)
                if (!readme) {
                    return null
                }
                return graphqlClient
                    .query({
                        query: TreePageReadmeQuery,
                        variables: {
                            repoID: resolvedRevision.repo.id,
                            revspec: resolvedRevision.commitID,
                            path: readme.path,
                        },
                    })
                    .then(result => {
                        if (result.data.node?.__typename !== 'Repository') {
                            throw new Error('Expected Repository')
                        }
                        return result.data.node.commit?.blob ?? null
                    })
            }),
        },
    }
}
