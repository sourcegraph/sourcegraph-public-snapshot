import * as Monaco from 'monaco-editor'
import { Token, toMonacoRange } from './scanner'
import { resolveFilter } from './filters'

/**
 * Returns the hover result for a hovered search token in the Monaco query input.
 */
export const getHoverResult = (
    tokens: Token[],
    { column }: Pick<Monaco.Position, 'column'>
): Monaco.languages.Hover | null => {
    const tokensAtCursor = tokens.filter(({ range }) => range.start + 1 <= column && range.end >= column)
    if (tokensAtCursor.length === 0) {
        return null
    }
    const values: string[] = []
    let range: Monaco.IRange | undefined
    tokensAtCursor.map(token => {
        switch (token.type) {
            case 'filter': {
                const resolvedFilter = resolveFilter(token.field.value)
                if (resolvedFilter) {
                    values.push(
                        'negated' in resolvedFilter
                            ? resolvedFilter.definition.description(resolvedFilter.negated)
                            : resolvedFilter.definition.description
                    )
                    range = toMonacoRange(token.range)
                }
                break
            }
        }
    })
    return {
        contents: values.map<Monaco.IMarkdownString>(
            (value): Monaco.IMarkdownString => ({
                value,
            })
        ),
        range,
    }
}
