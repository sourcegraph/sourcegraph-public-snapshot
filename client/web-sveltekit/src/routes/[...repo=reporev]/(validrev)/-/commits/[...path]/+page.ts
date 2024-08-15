import { redirect } from '@sveltejs/kit'

import { IncrementalRestoreStrategy, getGraphQLClient, infinityQuery } from '$lib/graphql'
import { resolveRevision } from '$lib/repo/utils'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { CommitsPage_CommitsQuery } from './page.gql'

const PAGE_SIZE = 20

export const load: PageLoad = async ({ parent, params, url }) => {
    const client = getGraphQLClient()
    const { repoName, revision = '' } = parseRepoRevision(params.repo)
    const path = params.path ? decodeURIComponent(params.path) : ''
    const resolvedRevision = resolveRevision(parent, revision)
    const isPerforceDepot = await parent().then(p => p.isPerforceDepot)

    if (isPerforceDepot) {
        const redirectURL = new URL(url)
        const pathItems = redirectURL.pathname.split('/')
        pathItems[pathItems.length - 1] = 'changelists'
        redirectURL.pathname = pathItems.join('/')
        redirect(301, redirectURL)
    }

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
        map: result => {
            const ancestors = result.data?.repository?.commit?.ancestors
            return {
                nextVariables:
                    ancestors?.pageInfo?.endCursor && ancestors.pageInfo.hasNextPage
                        ? { afterCursor: ancestors.pageInfo.endCursor }
                        : undefined,
                data: ancestors?.nodes,
                error: result.error,
            }
        },
        merge: (previous, next) => (previous ?? []).concat(next ?? []),
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
