import { findReadme } from '$lib/repo/tree'

import type { PageLoad } from './$types'
import { RepoPageReadmeQuery } from './page.gql'

export const load: PageLoad = async ({ parent }) => {
    const { resolvedRevision, deferred, graphqlClient } = await parent()

    return {
        deferred: {
            ...deferred,
            readme: deferred.fileTree.then(result => {
                const readme = findReadme(result.root.entries)
                if (!readme) {
                    return null
                }
                return graphqlClient
                    .query({
                        query: RepoPageReadmeQuery,
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
