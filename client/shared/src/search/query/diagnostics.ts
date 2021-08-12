import * as Monaco from 'monaco-editor'

import { SearchPatternType } from '../../graphql-operations'

import { validateFilter } from './filters'
import { toMonacoSingleLineRange } from './monaco'
import { PatternOf, each, matchesValue, eachOf } from './patternMatcher'
import { Filter, Token } from './token'

type FilterCheck = (f: Filter) => Monaco.editor.IMarkerData[]

export function validFilterValue(filter: Filter): Monaco.editor.IMarkerData[] {
    const validationResult = validateFilter(filter.field.value, filter.value)
    if (validationResult.valid) {
        return []
    }
    return [
        {
            severity: Monaco.MarkerSeverity.Error,
            message: `Error: ${validationResult.reason}`,
            ...toMonacoSingleLineRange(filter.field.range),
        },
    ]
}

export function emptyFilterValue(filter: Filter): Monaco.editor.IMarkerData[] {
    if (filter.value?.value !== '') {
        return []
    }
    return [
        {
            severity: Monaco.MarkerSeverity.Warning,
            message: `Warning: This filter is empty. Remove the space between the filter and value or quote the value to include the space. E.g., ${filter.field.value}:" a term".`,
            ...toMonacoSingleLineRange(filter.field.range),
        },
    ]
}

// Returns the first nonempty diagnostic for a filter, or nothing otherwise. We return
// the only the first so that we don't overwhelm the the user with multiple diagnostics
// for a single filter.
export function checkFilter(filter: Filter): Monaco.editor.IMarkerData[] {
    const checks: FilterCheck[] = [validFilterValue, emptyFilterValue]
    return checks.map(check => check(filter)).find(value => value.length !== 0) || []
}
const rules: PatternOf<Token[], Monaco.editor.IMarkerData[]>[] = [
    // Validate the value of each filter
    each({
        type: 'filter',
        $data: (token, context) => {
            context.data.push(...checkFilter(token as Filter))
        },
    }),
]

/**
 * Returns the diagnostics for a scanned search query to be displayed in the Monaco query input.
 */
export function getDiagnostics(tokens: Token[], patternType: SearchPatternType): Monaco.editor.IMarkerData[] {
    const result = matchesValue<Token[], Monaco.editor.IMarkerData[]>(tokens, eachOf(...rules), [])
    if (result.success) {
        return result.data
    }
    return []
}
