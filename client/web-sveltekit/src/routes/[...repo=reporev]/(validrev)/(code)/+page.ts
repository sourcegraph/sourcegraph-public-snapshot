import { findReadme } from '$lib/repo/tree'

import type { PageLoad } from './$types'
import { RepoPageReadmeQuery } from './page.gql'

export const load: PageLoad = async ({ parent }) => {
    const { resolvedRevision, graphqlClient, fileTree } = await parent()

    return {
        readme: fileTree.then(result => {
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
                        // This page will never render when the repository is not found.
                        // The (validrev) data loader will render an error page instead.
                        // Still, this error will show up as an unhandled promise rejection
                        // in the console. We should find a better way to handle this.
                        throw new Error('Expected Repository')
                    }
                    return result.data.node.commit?.blob ?? null
                })
        }),
    }
}
