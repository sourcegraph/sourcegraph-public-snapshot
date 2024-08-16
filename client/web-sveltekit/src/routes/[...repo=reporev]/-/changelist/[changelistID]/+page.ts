import { error } from '@sveltejs/kit'

import { IncrementalRestoreStrategy, getGraphQLClient, infinityQuery } from '$lib/graphql'

import type { PageLoad } from './$types'
import { ChangelistPage_ChangelistQuery, ChangelistPage_DiffQuery } from './page.gql'

const PAGE_SIZE = 20

export const load: PageLoad = async ({ params }) => {
    const client = getGraphQLClient()

    const result = await client.query(ChangelistPage_ChangelistQuery, {
        repoName: params.repo,
        cid: params.changelistID,
    })

    if (result.error) {
        error(500, `Unable to load commit data: ${result.error}`)
    }

    const changelist = result.data?.repository?.changelist

    if (!changelist) {
        error(404, 'Changelist not found')
    }

    // parents is an empty array for the initial commit
    // We currently don't support diffs for the initial commit on the backend

    const diff =
        changelist.cid && changelist?.commit.parents[0]?.parent?.cid
            ? infinityQuery({
                  client,
                  query: ChangelistPage_DiffQuery,
                  variables: {
                      repoName: params.repo,
                      base: changelist.commit.parents[0].oid,
                      head: changelist.commit.oid,
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
        changelist,
        diff,
    }
}
