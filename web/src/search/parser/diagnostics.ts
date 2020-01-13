import * as Monaco from 'monaco-editor'
import { Sequence } from './parser'
import { validateFilter } from './filters'

/**
 * Returns the diagnostics for a parsed search query to be displayed in the Monaco query input.
 */
export function getDiagnostics({ members }: Pick<Sequence, 'members'>): Monaco.editor.IMarkerData[] {
    const diagnostics: Monaco.editor.IMarkerData[] = []
    for (const { token, range } of members) {
        if (token.type === 'filter') {
            const { filterType, filterValue } = token
            const validationResult = validateFilter(filterType.token.value, filterValue)
            if (validationResult.valid) {
                continue
            }
            diagnostics.push({
                severity: Monaco.MarkerSeverity.Error,
                message: validationResult.reason,
                startLineNumber: 1,
                endLineNumber: 1,
                startColumn: filterType.range.start + 1,
                endColumn: filterType.range.end + 2,
            })
        } else if (token.type === 'literal') {
            if (token.value.includes(':')) {
                diagnostics.push({
                    severity: Monaco.MarkerSeverity.Warning,
                    message: 'Quoting the query may help if you want a literal match.',
                    startLineNumber: 1,
                    endLineNumber: 1,
                    startColumn: range.start + 1,
                    endColumn: range.end + 2,
                })
            }
        }
    }
    return diagnostics
}
