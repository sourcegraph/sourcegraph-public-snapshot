import { parseRepoRevision } from '@sourcegraph/shared/src/util/url'

import { getGraphQLClient, infinityQuery, mapOrThrow, IncrementalRestoreStrategy } from '$lib/graphql'

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
    // searchParams.get(...) can return null (hence the `?? 1`)
    // but its value might also not be convertible to a number
    // hence the `|| 1` to default to 1.
    const commitPage = +(url.searchParams.get('p') ?? 1) || 1

    return {
        commits: client
            .query(RepositoryComparisonCommits, {
                repoName,
                base,
                head,
                first: commitPage * PAGE_SIZE,
                path: filePath,
            })
            .then(
                mapOrThrow(result => ({
                    commits: (result.data?.repository?.comparison.commits.nodes || []).slice(
                        (commitPage - 1) * PAGE_SIZE,
                        commitPage * PAGE_SIZE
                    ),
                    nextPage: result.data?.repository?.comparison.commits.pageInfo.hasNextPage ? commitPage + 1 : null,
                    previousPage: commitPage > 1 ? commitPage - 1 : null,
                }))
            ),
        diffQuery: infinityQuery({
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
