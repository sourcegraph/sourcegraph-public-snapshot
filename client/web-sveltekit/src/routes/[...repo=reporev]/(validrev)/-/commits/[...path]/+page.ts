import { from } from 'rxjs'

import { getGraphQLClient, infinityQuery } from '$lib/graphql'
import { resolveRevision } from '$lib/repo/utils'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { CommitsPage_CommitsQuery } from './page.gql'

const PAGE_SIZE = 20

export const load: PageLoad = ({ parent, params }) => {
    const client = getGraphQLClient()
    const { repoName, revision = '' } = parseRepoRevision(params.repo)
    const path = params.path ? decodeURIComponent(params.path) : ''
    const resolvedRevision = resolveRevision(parent, revision)

    const commitsQuery = infinityQuery({
        client,
        query: CommitsPage_CommitsQuery,
        variables: from(
            resolvedRevision.then(revision => ({
                repoName,
                revision,
                first: PAGE_SIZE,
                path,
                afterCursor: null as string | null,
            }))
        ),
        nextVariables: previousResult => {
            if (previousResult?.data?.repository?.commit?.ancestors?.pageInfo?.hasNextPage) {
                return {
                    afterCursor: previousResult.data.repository.commit.ancestors.pageInfo.endCursor,
                }
            }
            return undefined
        },
        combine: (previousResult, nextResult) => {
            if (!nextResult.data?.repository?.commit) {
                return nextResult
            }
            const previousNodes = previousResult.data?.repository?.commit?.ancestors?.nodes ?? []
            const nextNodes = nextResult.data.repository?.commit?.ancestors.nodes ?? []
            return {
                ...nextResult,
                data: {
                    repository: {
                        ...nextResult.data.repository,
                        commit: {
                            ...nextResult.data.repository.commit,
                            ancestors: {
                                ...nextResult.data.repository.commit.ancestors,
                                nodes: [...previousNodes, ...nextNodes],
                            },
                        },
                    },
                },
            }
        },
    })

    return {
        commitsQuery,
        path,
    }
}
