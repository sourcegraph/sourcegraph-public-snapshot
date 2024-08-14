import { getGraphQLClient, mapOrThrow } from '$lib/graphql'
import { fetchTreeEntries } from '$lib/repo/api/tree'
import { findReadme } from '$lib/repo/tree'
import { resolveRevision } from '$lib/repo/utils'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { TreePageCommitInfoQuery, TreePageReadmeQuery } from './page.gql'

export const load: PageLoad = async ({ parent, params }) => {
    const client = getGraphQLClient()
    const { repoName, revision = '' } = parseRepoRevision(params.repo)
    const resolvedRevision = resolveRevision(parent, revision)
    const filePath = decodeURIComponent(params.path)

    const treeEntries = resolvedRevision
        .then(resolvedRevision =>
            fetchTreeEntries({
                repoName,
                revision: resolvedRevision,
                filePath,
                first: null,
            })
        )
        .then(commit => commit.tree)

    return {
        filePath,
        treeEntries,
        treeEntriesWithCommitInfo: resolvedRevision
            .then(resolvedRevision =>
                client.query(TreePageCommitInfoQuery, {
                    repoName,
                    revision: resolvedRevision,
                    filePath,
                    first: null,
                })
            )
            .then(
                mapOrThrow(result => {
                    if (!result.data?.repository) {
                        throw new Error('Unable to fetch repository information')
                    }
                    if (!result.data.repository.commit) {
                        throw new Error('Unable to fetch commit information')
                    }
                    return result.data.repository.commit.tree?.entries ?? []
                })
            ),
        readme: treeEntries.then(result => {
            if (!result) {
                return null
            }
            const readme = findReadme(result.entries)
            if (!readme) {
                return null
            }
            return resolvedRevision
                .then(resolvedRevision =>
                    client.query(TreePageReadmeQuery, {
                        repoName,
                        revision: resolvedRevision,
                        path: readme.path,
                    })
                )
                .then(
                    mapOrThrow(result => {
                        if (!result.data?.repository) {
                            throw new Error('Unable to fetch repository information')
                        }
                        if (!result.data.repository.commit) {
                            throw new Error('Unable to fetch commit information')
                        }
                        return result.data.repository.commit.blob
                    })
                )
        }),
    }
}
