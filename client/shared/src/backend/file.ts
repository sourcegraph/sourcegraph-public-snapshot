import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError, memoizeObservable } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { FetchFileParameters } from '@sourcegraph/search-ui'
import * as GQL from '@sourcegraph/shared/src/schema'

import {
    FormattedFileResult,
    FormattedFileVariables,
    HighlightedFileResult,
    HighlightedFileVariables,
} from '../graphql-operations'
import { PlatformContext } from '../platform/context'
import { makeRepoURI } from '../util/url'

export const fetchFormattedFileLineRanges = memoizeObservable(
    ({
        platformContext,
        ...context
    }: FetchFormattedFileParameters & {
        platformContext: Pick<PlatformContext, 'requestGraphQL'>
    }): Observable<string[][]> =>
        platformContext
            .requestGraphQL<FormattedFileResult, FormattedFileVariables>({
                request: gql`
                    query FormattedFile(
                        $repoName: String!
                        $commitID: String!
                        $filePath: String!
                        $ranges: [HighlightLineRange!]!
                    ) {
                        repository(name: $repoName) {
                            commit(rev: $commitID) {
                                file(path: $filePath) {
                                    isDirectory
                                    richHTML
                                    format {
                                        html
                                        lineRanges(ranges: $ranges)
                                    }
                                }
                            }
                        }
                    }
                `,
                variables: context,
                mightContainPrivateInfo: true,
            })
            .pipe(
                map(({ data, errors }) => {
                    if (!data?.repository?.commit?.file?.format) {
                        throw createAggregateError(errors)
                    }
                    const file = data.repository.commit.file
                    if (file.isDirectory) {
                        // TODO return []
                        return ''
                    }
                    return file.format.lineRanges
                })
            ),
    context =>
        makeRepoURI(context) + `?ranges=${context.ranges.map(range => `${range.startLine}:${range.endLine}`).join(',')}`
)

/**
 * Fetches the specified highlighted file line ranges (`FetchFileParameters.ranges`) and returns
 * them as a list of ranges, each describing a list of lines in the form of HTML table '<tr>...</tr>'.
 */
export const fetchHighlightedFileLineRanges = memoizeObservable(
    (
        {
            platformContext,
            ...context
        }: FetchFileParameters & {
            platformContext: Pick<PlatformContext, 'requestGraphQL'>
        },
        force?: boolean
    ): Observable<string[][]> =>
        platformContext
            .requestGraphQL<HighlightedFileResult, HighlightedFileVariables>({
                request: gql`
                    query HighlightedFile(
                        $repoName: String!
                        $commitID: String!
                        $filePath: String!
                        $disableTimeout: Boolean!
                        $ranges: [HighlightLineRange!]!
                    ) {
                        repository(name: $repoName) {
                            commit(rev: $commitID) {
                                file(path: $filePath) {
                                    isDirectory
                                    richHTML
                                    highlight(disableTimeout: $disableTimeout) {
                                        aborted
                                        lineRanges(ranges: $ranges)
                                    }
                                }
                            }
                        }
                    }
                `,
                variables: { ...context, disableTimeout: !!context.disableTimeout },
                mightContainPrivateInfo: true,
            })
            .pipe(
                map(({ data, errors }) => {
                    if (!data?.repository?.commit?.file?.highlight) {
                        throw createAggregateError(errors)
                    }
                    const file = data.repository.commit.file
                    if (file.isDirectory) {
                        return []
                    }
                    return file.highlight.lineRanges
                })
            ),
    context =>
        makeRepoURI(context) +
        `?disableTimeout=${String(context.disableTimeout)}&ranges=${context.ranges
            .map(range => `${range.startLine}:${range.endLine}`)
            .join(',')}`
)

export interface FetchFormattedFileParameters {
    repoName: string
    commitID: string
    filePath: string
    ranges: GQL.IHighlightLineRange[]
}
