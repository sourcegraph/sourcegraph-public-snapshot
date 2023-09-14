import { gql, query } from '$lib/graphql'
import {
    HighlightResponseFormat,
    type HighlightLineRange,
    type HighlightedFileResult,
    type HighlightedFileVariables,
} from '$lib/graphql-operations'

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
            id
            commit(rev: $commitID) {
                id
                blob(path: $filePath) {
                    canonicalURL
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
export const fetchFileRangeMatches = async (args: {
    result: Result
    format?: HighlightResponseFormat
    ranges: HighlightLineRange[]
}): Promise<string[][]> => {
    const data = await query<HighlightedFileResult, HighlightedFileVariables>(HIGHLIGHTED_FILE_QUERY, {
        repoName: args.result.repository,
        commitID: args.result.commit ?? '',
        filePath: args.result.path,
        ranges: args.ranges,
        format: args.format ?? HighlightResponseFormat.HTML_HIGHLIGHT,
        disableTimeout: true,
    })

    if (!data?.repository?.commit?.blob?.highlight) {
        throw new Error('Unable to highlight file range')
    }

    const file = data.repository.commit.blob
    if (file?.isDirectory) {
        return []
    }

    return file?.highlight.lineRanges ?? []
}
