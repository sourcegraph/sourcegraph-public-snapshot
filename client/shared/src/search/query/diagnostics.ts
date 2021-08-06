import * as Monaco from 'monaco-editor'

import { SearchPatternType } from '../../graphql-operations'

import { validateFilter } from './filters'
import { toMonacoSingleLineRange } from './monaco'
import { Filter, Token } from './token'

export function checkValidValue(filter: Filter): Monaco.editor.IMarkerData[] {
    const validationResult = validateFilter(filter.field.value, filter.value)
    if (validationResult.valid) {
        return []
    }
    return [
        {
            severity: Monaco.MarkerSeverity.Error,
            message: validationResult.reason,
            ...toMonacoSingleLineRange(filter.field.range),
        },
    ]
}

/**
 * Returns the diagnostics for a scanned search query to be displayed in the Monaco query input.
 */
export function getDiagnostics(tokens: Token[], patternType: SearchPatternType): Monaco.editor.IMarkerData[] {
    return tokens.filter(token => token.type === 'filter').flatMap(filter => checkValidValue(filter as Filter))
}
