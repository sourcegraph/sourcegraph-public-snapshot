import { getGraphQLClient, mapOrThrow } from '$lib/graphql'
import { findReadme } from '$lib/repo/tree'

import type { PageLoad } from './$types'
import { RepoPageReadmeQuery } from './page.gql'

export const load: PageLoad = ({ parent }) => {
    return {
        readme: parent().then(({ resolvedRevision, fileTree, repoName }) =>
            fileTree
                .then(result => {
                    if (!result) {
                        return null
                    }
                    const readme = findReadme(result.root.entries)
                    if (!readme) {
                        return null
                    }
                    return getGraphQLClient()
                        .query(RepoPageReadmeQuery, {
                            repoName,
                            revspec: resolvedRevision.commitID,
                            path: readme.path,
                        })
                        .then(mapOrThrow(result => result.data?.repository?.commit?.blob ?? null))
                })
                .catch(() => null)
        ),
    }
}
