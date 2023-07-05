import type { Observable } from 'rxjs'
import { get } from 'svelte/store'

import type { HighlightResponseFormat, HighlightLineRange } from '$lib/graphql-operations'
import { fetchHighlightedFileLineRanges } from '$lib/loader/blob'
import { platformContext } from '$lib/stores'

interface Result {
    repository: string
    commit?: string
    path: string
}

export function fetchFileRangeMatches(args: {
    result: Result
    format?: HighlightResponseFormat
    ranges: HighlightLineRange[]
}): Observable<string[][]> {
    return fetchHighlightedFileLineRanges(
        {
            repoName: args.result.repository,
            commitID: args.result.commit || '',
            filePath: args.result.path,
            disableTimeout: false,
            format: args.format,
            ranges: args.ranges,
            platformContext: get(platformContext),
        },
        false
    )
}
