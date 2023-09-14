import { addLineRangeQueryParameter, formatSearchParameters, toPositionOrRangeQueryParameter } from '$lib/common'
import type { SelectedLineRange } from '$lib/web'

export function updateSearchParamsWithLineInformation(
    currentSearchParams: URLSearchParams,
    range: SelectedLineRange
): string {
    const parameters = new URLSearchParams(currentSearchParams)
    parameters.delete('popover')

    let query: string | undefined

    if (range?.line !== range?.endLine && range?.endLine) {
        query = toPositionOrRangeQueryParameter({
            range: {
                start: { line: range.line },
                end: { line: range.endLine },
            },
        })
    } else if (range?.line) {
        query = toPositionOrRangeQueryParameter({ position: { line: range.line } })
    }

    return formatSearchParameters(addLineRangeQueryParameter(parameters, query))
}
