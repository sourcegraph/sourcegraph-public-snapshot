import * as Monaco from 'monaco-editor'
import { Observable, fromEventPattern, of } from 'rxjs'
import { map, takeUntil, switchMap, delay } from 'rxjs/operators'

import { SearchPatternType } from '../../graphql-operations'
import { SearchMatch } from '../stream'

import { getCompletionItems } from './completion'
import { getMonacoTokens } from './decoratedToken'
import { getHoverResult } from './hover'
import { scanSearchQuery } from './scanner'
import { isRepoFilter } from './validate'

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

/**
 * Returns the providers used by the Monaco query input to provide syntax highlighting,
 * hovers and completions for the Sourcegraph search syntax.
 */
export function getProviders(
    fetchSuggestions: (input: string) => Observable<SearchMatch[]>,
    options: {
        patternType: SearchPatternType
        globbing: boolean
        interpretComments?: boolean
        isSourcegraphDotCom?: boolean
    }
): SearchFieldProviders {
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
            provideCompletionItems: (textModel, position, context, cancellationToken) => {
                const value = textModel.getValue()

                const scanned = scanSearchQuery(value, options.interpretComments ?? false, options.patternType)
                if (scanned.type === 'error') {
                    return null
                }

                const tokenAtColumn = scanned.term.find(
                    ({ range }) => range.start + 1 <= position.column && range.end + 1 >= position.column
                )
                if (!tokenAtColumn) {
                    throw new Error('getCompletionItems: no token at column')
                }

                if (isRepoFilter(tokenAtColumn)) {
                    const suggestionQuery = `${tokenAtColumn.value?.value ?? ''} type:repo patterntype:regexp count:50`

                    return of(suggestionQuery)
                        .pipe(
                            // We use a delay here to implement a custom debounce. In the next step we check if the current
                            // completion request was cancelled in the meantime (`token.isCancellationRequested`).
                            // This prevents us from needlessly running multiple suggestion queries.
                            delay(100),
                            switchMap(query =>
                                cancellationToken.isCancellationRequested
                                    ? Promise.resolve(null)
                                    : getCompletionItems(
                                          tokenAtColumn,
                                          position,
                                          fetchSuggestions(query),
                                          options.globbing,
                                          options.isSourcegraphDotCom
                                      )
                            )
                        )
                        .toPromise()
                }

                return null
            },
        },
    }
}
