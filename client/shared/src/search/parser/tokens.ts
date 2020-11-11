import * as Monaco from 'monaco-editor'
import { Token } from './scanner'

/**
 * Returns the tokens in a scanned search query displayed in the Monaco query input.
 */
export function getMonacoTokens(tokens: Token[]): Monaco.languages.IToken[] {
    const monacoTokens: Monaco.languages.IToken[] = []
    for (const token of tokens) {
        switch (token.type) {
            case 'filter':
                {
                    monacoTokens.push({
                        startIndex: token.filterType.range.start,
                        scopes: 'filterKeyword',
                    })
                    if (token.filterValue) {
                        monacoTokens.push({
                            startIndex: token.filterValue.range.start,
                            scopes: 'identifier',
                        })
                    }
                }
                break
            case 'whitespace':
            case 'keyword':
            case 'comment':
                monacoTokens.push({
                    startIndex: token.range.start,
                    scopes: token.type,
                })
                break
            default:
                monacoTokens.push({
                    startIndex: token.range.start,
                    scopes: 'identifier',
                })
                break
        }
    }
    return monacoTokens
}
