import * as Monaco from 'monaco-editor'
import { Sequence } from './scanner'

/**
 * Returns the tokens in a scanned search query displayed in the Monaco query input.
 */
export function getMonacoTokens(scannedQuery: Pick<Sequence, 'members'>): Monaco.languages.IToken[] {
    const tokens: Monaco.languages.IToken[] = []
    for (const token of scannedQuery.members) {
        switch (token.type) {
            case 'filter':
                {
                    tokens.push({
                        startIndex: token.filterType.range.start,
                        scopes: 'filterKeyword',
                    })
                    if (token.filterValue) {
                        tokens.push({
                            startIndex: token.filterValue.range.start,
                            scopes: 'identifier',
                        })
                    }
                }
                break
            case 'whitespace':
            case 'keyword':
            case 'comment':
                tokens.push({
                    startIndex: token.range.start,
                    scopes: token.type,
                })
                break
            default:
                tokens.push({
                    startIndex: token.range.start,
                    scopes: 'identifier',
                })
                break
        }
    }
    return tokens
}
