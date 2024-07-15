import { IncrementalRestoreStrategy, getGraphQLClient, infinityQuery } from '$lib/graphql'
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
        variables: resolvedRevision.then(revision => ({
            repoName,
            revision,
            first: PAGE_SIZE,
            path,
            afterCursor: null as string | null,
        })),
        mapResult(result, previousResult) {
            const ancestors = result.data?.repository?.commit?.ancestors
            return {
                nextVariables:
                    ancestors?.pageInfo?.endCursor && ancestors.pageInfo.hasNextPage
                        ? { afterCursor: ancestors.pageInfo.endCursor }
                        : undefined,
                data: (previousResult.data ?? []).concat(ancestors?.nodes ?? []),
                error: result.error,
            }
        },
        createRestoreStrategy: api =>
            new IncrementalRestoreStrategy(
                api,
                n => n.length,
                n => ({ first: n.length })
            ),
    })

    return {
        commitsQuery,
        path,
    }
}
