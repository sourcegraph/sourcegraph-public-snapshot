import * as Monaco from 'monaco-editor'
import { Sequence } from './parser'
import { FILTERS } from './filters'

/**
 * Returns the hover result for a hovered search token in the Monaco query input.
 */
export const getHoverResult = (
    { members }: Pick<Sequence, 'members'>,
    { column }: Pick<Monaco.Position, 'column'>
): Monaco.languages.Hover | null => {
    const tokenAtColumn = members.find(({ range }) => range.start + 1 <= column && range.end + 1 >= column)
    if (!tokenAtColumn || tokenAtColumn.token.type !== 'filter') {
        return null
    }
    const { filterType } = tokenAtColumn.token
    const matchedFilterDefinition = FILTERS.find(({ aliases }) => aliases.includes(filterType.token.value))
    if (!matchedFilterDefinition) {
        return null
    }
    return {
        contents: [{ value: matchedFilterDefinition.description }],
        range: {
            startLineNumber: 1,
            endLineNumber: 1,
            startColumn: tokenAtColumn.range.start + 1,
            endColumn: tokenAtColumn.range.end + 2,
        },
    }
}
