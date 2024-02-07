import { getGraphQLClient } from '$lib/graphql'
import { fetchTreeEntries } from '$lib/repo/api/tree'
import { findReadme } from '$lib/repo/tree'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { TreePageCommitInfoQuery, TreePageReadmeQuery } from './page.gql'

export const load: PageLoad = async ({ params }) => {
    const client = await getGraphQLClient()
    const { repoName, revision = '' } = parseRepoRevision(params.repo)

    const treeEntries = fetchTreeEntries({
        repoName,
        revision,
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
                    revision,
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
                        revision,
                        path: readme.path,
                    },
                })
                .then(result => {
                    return result.data.repository?.commit?.blob ?? null
                })
        }),
    }
}
