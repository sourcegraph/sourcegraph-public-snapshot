import { getGraphQLClient } from '$lib/graphql'
import { fetchTreeEntries } from '$lib/repo/api/tree'
import { findReadme } from '$lib/repo/tree'
import { resolveRevision } from '$lib/repo/utils'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { TreePageCommitInfoQuery, TreePageReadmeQuery } from './page.gql'

export const load: PageLoad = async ({ parent, params }) => {
    const client = await getGraphQLClient()
    const { repoName, revision = '' } = parseRepoRevision(params.repo)
    const resolvedRevision = await resolveRevision(parent, revision)

    const treeEntries = fetchTreeEntries({
        repoName,
        revision: resolvedRevision,
        filePath: params.path,
        first: null,
    }).then(
        commit => commit.tree,
        () => null
    )

    return {
        filePath: params.path,
        treeEntries,
        commitInfo: client
            .query({
                query: TreePageCommitInfoQuery,
                variables: {
                    repoName,
                    revision: resolvedRevision,
                    filePath: params.path,
                    first: null,
                },
            })
            .then(result => {
                return result.data.repository?.commit?.tree ?? null
            }),
        readme: treeEntries.then(result => {
            if (!result) {
                return null
            }
            const readme = findReadme(result.entries)
            if (!readme) {
                return null
            }
            return client
                .query({
                    query: TreePageReadmeQuery,
                    variables: {
                        repoName,
                        revision: resolvedRevision,
                        path: readme.path,
                    },
                })
                .then(result => {
                    return result.data.repository?.commit?.blob ?? null
                })
        }),
    }
}
