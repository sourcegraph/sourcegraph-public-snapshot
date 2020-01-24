import * as Monaco from 'monaco-editor'
import { Sequence } from './parser'

/**
 * Returns the tokens in a parsed search query displayed in the Monaco query input.
 */
export function getMonacoTokens(parsedQuery: Pick<Sequence, 'members'>): Monaco.languages.IToken[] {
    const tokens: Monaco.languages.IToken[] = []
    for (const { token, range } of parsedQuery.members) {
        if (token.type === 'whitespace') {
            tokens.push({
                startIndex: range.start,
                scopes: 'whitespace',
            })
        } else if (token.type === 'quoted' || token.type === 'literal') {
            tokens.push({
                startIndex: range.start,
                scopes: 'identifier',
            })
        } else if (token.type === 'filter') {
            tokens.push({
                startIndex: token.filterType.range.start,
                scopes: 'keyword',
            })
            if (token.filterValue) {
                tokens.push({
                    startIndex: token.filterValue.range.start,
                    scopes: 'identifier',
                })
            }
        }
    }
    return tokens
}
