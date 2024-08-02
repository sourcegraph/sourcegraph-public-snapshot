import { error } from '@sveltejs/kit'

import { IncrementalRestoreStrategy, getGraphQLClient, infinityQuery } from '$lib/graphql'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { CommitPage_CommitQuery, CommitPage_DiffQuery, CommitPage_Changelist } from './page.gql'

const PAGE_SIZE = 20

export const load: PageLoad = async ({ parent, params }) => {
    const client = getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)
    const { resolvedRepository } = await parent()

    const isPerforceDepot = resolvedRepository.externalRepository.serviceType === 'perforce'

    const result = isPerforceDepot
        ? await client.query(CommitPage_Changelist, { repoId: resolvedRepository.id, changelistId: params.revspec })
        : await client.query(CommitPage_CommitQuery, { repoName, revspec: params.revspec })

    if (result.error) {
        error(500, `Unable to load commit data: ${result.error}`)
    }

    const commit = 'node' in result?.data ? result.data?.node?.changelist.commit : result.data?.repository?.commit

    if (!commit) {
        error(404, 'Commit not found')
    }

    // parents is an empty array for the initial commit
    // We currently don't support diffs for the initial commit on the backend
    const diff =
        commit?.oid && commit?.parents[0]?.oid
            ? infinityQuery({
                  client,
                  query: CommitPage_DiffQuery,
                  variables: {
                      repoName,
                      base: commit.parents[0].oid,
                      head: commit.oid,
                      first: PAGE_SIZE,
                      after: null as string | null,
                  },
                  map: result => {
                      const diffs = result.data?.repository?.comparison.fileDiffs
                      return {
                          nextVariables: diffs?.pageInfo.hasNextPage ? { after: diffs?.pageInfo.endCursor } : undefined,
                          data: diffs?.nodes,
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
            : null

    return {
        commit,
        diff,
    }
}
