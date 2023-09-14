import type * as Monaco from 'monaco-editor'
import { type Observable, fromEventPattern, of } from 'rxjs'
import { map, takeUntil } from 'rxjs/operators'

import type { SearchPatternType } from '../../graphql-operations'
import { isSearchMatchOfType, type SearchMatch } from '../stream'

import { getCompletionItems } from './completion'
import { getMonacoTokens } from './decoratedToken'
import { getHoverResult } from './hover'
import { createCancelableFetchSuggestions, getSuggestionQuery } from './providers-utils'
import { scanSearchQuery } from './scanner'
import type { Token } from './token'

interface SearchFieldProviders {
    tokens: Monaco.languages.TokensProvider
    hover: Monaco.languages.HoverProvider
    completion: Monaco.languages.CompletionItemProvider
}

/**
 * A dummy scanner state, required for the token provider.
 */
const SCANNER_STATE: Monaco.languages.IState = {
    clone: () => ({ ...SCANNER_STATE }),
    equals: () => false,
}

const printable = ' !"#$%&\'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~'
const latin1Alpha = 'ÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖ×ØÙÚÛÜÝÞßàáâãäåæçèéêëìíîïðñòóôõö÷øùúûüýþÿ'

function getTokenAtColumn(tokens: Token[], column: number): Token | null {
    return tokens.find(({ range }) => range.start + 1 <= column && range.end + 1 >= column) ?? null
}

/**
 * Returns the providers used by the Monaco query input to provide syntax highlighting,
 * hovers and completions for the Sourcegraph search syntax.
 */
export function getProviders(
    fetchSuggestions: (input: string) => Observable<SearchMatch[]>,
    options: {
        patternType: SearchPatternType
        disablePatternSuggestions?: boolean
        interpretComments?: boolean
        isSourcegraphDotCom?: boolean
    }
): SearchFieldProviders {
    const cancelableFetch = createCancelableFetchSuggestions(fetchSuggestions)

    return {
        tokens: {
            getInitialState: () => SCANNER_STATE,
            tokenize: line => {
                const result = scanSearchQuery(line, options.interpretComments ?? false, options.patternType)
                if (result.type === 'success') {
                    return {
                        tokens: getMonacoTokens(result.term),
                        endState: SCANNER_STATE,
                    }
                }
                return { endState: SCANNER_STATE, tokens: [] }
            },
        },
        hover: {
            provideHover: (textModel, position, token) =>
                of(textModel.getValue())
                    .pipe(
                        map(value => scanSearchQuery(value, options.interpretComments ?? false, options.patternType)),
                        map(scanned =>
                            scanned.type === 'error' ? null : getHoverResult(scanned.term, position, textModel)
                        ),
                        takeUntil(fromEventPattern(handler => token.onCancellationRequested(handler)))
                    )
                    .toPromise(),
        },
        completion: {
            // An explicit list of trigger characters is needed for the Monaco editor to show completions.
            triggerCharacters: [...printable, ...latin1Alpha],
            provideCompletionItems: (textModel, position, _context, cancellationToken) => {
                const scanned = scanSearchQuery(
                    textModel.getValue(),
                    options.interpretComments ?? false,
                    options.patternType
                )
                if (scanned.type === 'error') {
                    return null
                }
                const tokenAtColumn = getTokenAtColumn(scanned.term, position.column)

                return getCompletionItems(
                    tokenAtColumn,
                    position,
                    (token, type) =>
                        cancelableFetch(getSuggestionQuery(scanned.term, token, type), listener =>
                            cancellationToken.onCancellationRequested(listener)
                        ).then(matches => matches.filter(isSearchMatchOfType(type))),
                    options
                )
            },
        },
    }
}
