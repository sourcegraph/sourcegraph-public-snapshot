import type { Observable } from 'rxjs'

import type { HighlightResponseFormat, HighlightLineRange } from '$lib/graphql-operations'
import { fetchHighlightedFileLineRanges } from '$lib/loader/blob'
import type { PlatformContext } from '$lib/shared'

interface Result {
    repository: string
    commit?: string
    path: string
}

export function fetchFileRangeMatches(args: {
    result: Result
    format?: HighlightResponseFormat
    ranges: HighlightLineRange[]
    platformContext: PlatformContext
}): Observable<string[][]> {
    return fetchHighlightedFileLineRanges(
        {
            repoName: args.result.repository,
            commitID: args.result.commit || '',
            filePath: args.result.path,
            disableTimeout: false,
            format: args.format,
            ranges: args.ranges,
            platformContext: args.platformContext,
        },
        false
    )
}
