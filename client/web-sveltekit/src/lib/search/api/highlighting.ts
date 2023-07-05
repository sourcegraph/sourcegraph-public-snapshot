import { memoize } from 'lodash'

import { createAggregateError } from '$lib/common'
import {
    HighlightResponseFormat,
    type HighlightLineRange,
    type HighlightedFileResult,
    type HighlightedFileVariables,
} from '$lib/graphql-operations'
import { getDocumentNode, gql } from '$lib/http-client'
import { makeRepoURI } from '$lib/shared'
import { getWebGraphQLClient } from '$lib/web'

const HIGHLIGHTED_FILE_QUERY = gql`
    query HighlightedFile(
        $repoName: String!
        $commitID: String!
        $filePath: String!
        $disableTimeout: Boolean!
        $ranges: [HighlightLineRange!]!
        $format: HighlightResponseFormat!
    ) {
        repository(name: $repoName) {
            commit(rev: $commitID) {
                file(path: $filePath) {
                    isDirectory
                    highlight(disableTimeout: $disableTimeout, format: $format) {
                        aborted
                        lineRanges(ranges: $ranges)
                    }
                }
            }
        }
    }
`

interface Result {
    repository: string
    commit?: string
    path: string
}

/**
 * Fetches the specified highlighted file line ranges (`FetchFileParameters.ranges`) and returns
 * them as a list of ranges, each describing a list of lines in the form of HTML table '<tr>...</tr>'.
 */
export const fetchFileRangeMatches = memoize(
    async (args: {
        result: Result
        format?: HighlightResponseFormat
        ranges: HighlightLineRange[]
    }): Promise<string[][]> => {
        const client = await getWebGraphQLClient()
        const result = await client.query<HighlightedFileResult, HighlightedFileVariables>({
            query: getDocumentNode(HIGHLIGHTED_FILE_QUERY),
            variables: {
                repoName: args.result.repository,
                commitID: args.result.commit ?? '',
                filePath: args.result.path,
                ranges: args.ranges,
                format: args.format ?? HighlightResponseFormat.HTML_HIGHLIGHT,
                disableTimeout: true,
            },
        })

        if (!result.data?.repository?.commit?.file?.highlight) {
            throw createAggregateError(result.errors)
        }

        const file = result.data.repository.commit.file
        if (file.isDirectory) {
            return []
        }

        return file.highlight.lineRanges
    },
    ({ result: { repository, commit, path }, format, ranges }) =>
        makeRepoURI({ repoName: repository, commitID: commit, filePath: path }) +
        `?ranges=${ranges.map(range => `${range.startLine}:${range.endLine}`).join(',')}&format=${format}`
)
