import * as Monaco from 'monaco-editor'
import { Observable, fromEventPattern, of } from 'rxjs'
import { map, takeUntil, switchMap, delay } from 'rxjs/operators'

import { SearchPatternType } from '../../graphql-operations'
import { SearchMatch } from '../stream'

import { getCompletionItems } from './completion'
import { getMonacoTokens } from './decoratedToken'
import { FilterType } from './filters'
import { getHoverResult } from './hover'
import { scanSearchQuery } from './scanner'
import { Filter, KeywordKind, Token } from './token'
import { isFilterType } from './validate'

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

function collectFilterTokens(tokens: Token[], filterType: FilterType): Filter[] {
    return tokens.filter(token => isFilterType(token, filterType)) as Filter[]
}

function serializeFilterTokens(filters: Filter[]): string {
    return filters
        .map(filter => (filter.value ? `${filter.field.value}:${filter.value.value}` : ''))
        .filter(filter => !!filter)
        .join(' ')
}

const MAX_SUGGESTION_COUNT = 50

function getSuggestionQuery(tokens: Token[], tokenAtColumn: Token): string {
    const hasAndOrOperators = tokens.some(
        token => token.type === 'keyword' && (token.kind === KeywordKind.Or || token.kind === KeywordKind.And)
    )

    if (isFilterType(tokenAtColumn, FilterType.repo) && tokenAtColumn.value) {
        return `repo:${tokenAtColumn.value.value} type:repo count:${MAX_SUGGESTION_COUNT}`
    }
    if (isFilterType(tokenAtColumn, FilterType.file) && tokenAtColumn.value && !hasAndOrOperators) {
        const repoQueryPart = serializeFilterTokens(collectFilterTokens(tokens, FilterType.repo))
        return `${repoQueryPart} file:${tokenAtColumn.value.value} type:path count:${MAX_SUGGESTION_COUNT}`
    }
    if (tokenAtColumn.type === 'pattern' && tokenAtColumn.value && !hasAndOrOperators) {
        const repoQueryPart = serializeFilterTokens(collectFilterTokens(tokens, FilterType.repo))
        const fileQueryPart = serializeFilterTokens(collectFilterTokens(tokens, FilterType.file))
        return `${repoQueryPart} ${fileQueryPart} ${tokenAtColumn.value} type:symbol count:${MAX_SUGGESTION_COUNT}`
    }

    return ''
}

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
                    return null
                }

                return of(getSuggestionQuery(scanned.term, tokenAtColumn))
                    .pipe(
                        // We use a delay here to implement a custom debounce. In the next step we check if the current
                        // completion request was cancelled in the meantime (`token.isCancellationRequested`).
                        // This prevents us from needlessly running multiple suggestion queries.
                        delay(200),
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
            },
        },
    }
}
