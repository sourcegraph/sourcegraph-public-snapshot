import { parseRepoRevision } from '@sourcegraph/shared/src/util/url'

import {
    getGraphQLClient,
    mapOrThrow,
    IncrementalRestoreStrategy,
    createPagination,
    OverwriteRestoreStrategy,
} from '$lib/graphql'

import type { PageLoad } from './$types'
import { RepositoryComparison, RepositoryComparisonCommits, RepositoryComparisonDiffs } from './page.gql'

const PAGE_SIZE = 10

export const load: PageLoad = async ({ params, url }) => {
    const client = getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)
    const filePath = url.searchParams.get('filePath') ?? ''

    let baseRevspec = ''
    let headRevspec = ''

    if (params.rangeSpec.includes('...')) {
        ;[baseRevspec, headRevspec] = params.rangeSpec.split('...', 2)
    }

    if (!baseRevspec && !headRevspec) {
        return {
            head: null,
            base: null,
            error: null,
        }
    }

    const base = baseRevspec || null
    const head = headRevspec || null

    return {
        commitsPagination: createPagination({
            client,
            query: RepositoryComparisonCommits,
            variables: {
                repoName,
                base,
                head,
                first: PAGE_SIZE,
                path: filePath,
            },
            map: result => {
                // The API doesn't implement real pagination. Instead we fetch the first N and then show the last
                // PAGE_SIZE commits.
                const commits = result.data?.repository?.comparison?.commits
                const nodes = commits?.nodes ?? []
                const page = Math.ceil(nodes.length / PAGE_SIZE)
                const start = (page - 1) * PAGE_SIZE
                const end = page * PAGE_SIZE

                return {
                    nextVariables: commits?.pageInfo.hasNextPage ? { first: (page + 1) * PAGE_SIZE } : undefined,
                    prevVariables: page > 1 ? { first: (page - 1) * PAGE_SIZE } : undefined,
                    data: {
                        commits: nodes.slice(start, end),
                        // Used for restoring the state when navigating back
                        totalCount: nodes.length,
                    },
                    error: result.error,
                }
            },

            createRestoreStrategy(api) {
                return new OverwriteRestoreStrategy(api, data => ({ first: data.totalCount }))
            },
        }),
        diffPagination: createPagination({
            client,
            query: RepositoryComparisonDiffs,
            variables: {
                repoName,
                base,
                head,
                first: PAGE_SIZE,
                after: null as string | null,
                paths: filePath ? [filePath] : [],
            },
            map: result => {
                const fileDiffs = result.data?.repository?.comparison?.fileDiffs
                return {
                    nextVariables: fileDiffs?.pageInfo.hasNextPage
                        ? { after: fileDiffs.pageInfo.endCursor }
                        : undefined,
                    data: fileDiffs?.nodes,
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
        }),
        // We await this here so that we are not blocking the commits and diff queries
        ...(await client
            .query(RepositoryComparison, {
                repoName,
                base,
                head,
            })
            .toPromise()
            .then(
                mapOrThrow(result => {
                    const range = result.data?.repository?.comparison?.range
                    const baseCommitID = range?.baseRevSpec.object?.oid
                    const headCommitID = range?.headRevSpec.object?.oid

                    if (!baseCommitID) {
                        throw new Error(`Base revision '${baseRevspec}' not found`)
                    }
                    if (!headCommitID) {
                        throw new Error(`Head revision '${headRevspec}' not found`)
                    }

                    return {
                        base: {
                            revision: baseRevspec,
                            commitID: baseCommitID,
                        },
                        head: {
                            revision: headRevspec,
                            commitID: headCommitID,
                        },
                    }
                })
            )
            .catch(error => ({ error }))),
    }
}
