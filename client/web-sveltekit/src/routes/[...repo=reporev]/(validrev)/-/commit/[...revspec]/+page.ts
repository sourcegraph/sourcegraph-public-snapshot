import { error } from '@sveltejs/kit'

import { getGraphQLClient, infinityQuery } from '$lib/graphql'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { CommitPage_CommitQuery, CommitPage_DiffQuery } from './page.gql'

const PAGE_SIZE = 20

export const load: PageLoad = async ({ parent, params }) => {
    const client = getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)

    const result = await client.query(CommitPage_CommitQuery, { repoName, revspec: params.revspec })

    if (result.error) {
        error(500, `Unable to load commit data: ${result.error}`)
    }

    const commit = result.data?.repository?.commit

    if (!commit) {
        error(404, 'Commit not found')
    }

    parent().then(parent => parent.repositoryContext.set({ revision: commit.abbreviatedOID }))

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
                  nextVariables: previousResult => {
                      if (
                          !previousResult.error &&
                          previousResult?.data?.repository?.comparison?.fileDiffs?.pageInfo?.hasNextPage
                      ) {
                          return {
                              after: previousResult.data.repository.comparison.fileDiffs.pageInfo.endCursor,
                          }
                      }
                      return undefined
                  },
                  combine: (previousResult, nextResult) => {
                      if (!nextResult.data?.repository?.comparison) {
                          return {
                              ...nextResult,
                              // When this code path is executed we probably have an error.
                              // We still want to show the data that was loaded before the error occurred.
                              data: previousResult.data,
                          }
                      }
                      const previousNodes = previousResult.data?.repository?.comparison?.fileDiffs?.nodes ?? []
                      const nextNodes = nextResult.data.repository?.comparison?.fileDiffs?.nodes ?? []
                      return {
                          ...nextResult,
                          data: {
                              repository: {
                                  ...nextResult.data.repository,
                                  comparison: {
                                      ...nextResult.data.repository.comparison,
                                      fileDiffs: {
                                          ...nextResult.data.repository.comparison.fileDiffs,
                                          nodes: [...previousNodes, ...nextNodes],
                                      },
                                  },
                              },
                          },
                      }
                  },
              })
            : null

    return {
        commit,
        diff,
    }
}
