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

function serializeFilters(tokens: Token[], filterTypes: FilterType[]): string {
    return tokens
        .filter((token): token is Filter => filterTypes.some(filterType => isFilterType(token, filterType)))
        .map(filter => (filter.value ? `${filter.field.value}:${filter.value.value}` : ''))
        .filter(serialized => !!serialized)
        .join(' ')
}

const MAX_SUGGESTION_COUNT = 50
const REPO_SUGGESTION_FILTERS = [FilterType.fork, FilterType.visibility, FilterType.archived]
const FILE_SUGGESTION_FILTERS = [...REPO_SUGGESTION_FILTERS, FilterType.repo, FilterType.rev, FilterType.lang]

export function getSuggestionQuery(tokens: Token[], tokenAtColumn: Token): string {
    const hasAndOrOperators = tokens.some(
        token => token.type === 'keyword' && (token.kind === KeywordKind.Or || token.kind === KeywordKind.And)
    )

    if (isFilterType(tokenAtColumn, FilterType.repo) && tokenAtColumn.value) {
        const relevantFilters = !hasAndOrOperators ? serializeFilters(tokens, REPO_SUGGESTION_FILTERS) : ''
        return `${relevantFilters} repo:${tokenAtColumn.value.value} type:repo count:${MAX_SUGGESTION_COUNT}`.trimStart()
    }

    // For the cases below, we are not handling queries with and/or operators. This is because we would need to figure out
    // for each filter which filters from the surrounding expression apply to it. For example, if we have a query: `repo:x file:y z OR repo:xx file:yy`
    // and we want to get suggestions for the `file:yy` filter. We would only want to include file suggestions from the `xx` repo and not the `x` repo, because it
    // is a part of a different expression.
    if (hasAndOrOperators) {
        return ''
    }
    if (isFilterType(tokenAtColumn, FilterType.file) && tokenAtColumn.value) {
        const relevantFilters = serializeFilters(tokens, FILE_SUGGESTION_FILTERS)
        return `${relevantFilters} file:${tokenAtColumn.value.value} type:path count:${MAX_SUGGESTION_COUNT}`.trimStart()
    }
    if (tokenAtColumn.type === 'pattern' && tokenAtColumn.value) {
        const relevantFilters = serializeFilters(tokens, [...FILE_SUGGESTION_FILTERS, FilterType.file])
        return `${relevantFilters} ${tokenAtColumn.value} type:symbol count:${MAX_SUGGESTION_COUNT}`.trimStart()
    }

    return ''
}

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
                const scanned = scanSearchQuery(
                    textModel.getValue(),
                    options.interpretComments ?? false,
                    options.patternType
                )
                if (scanned.type === 'error') {
                    return null
                }
                const tokenAtColumn = getTokenAtColumn(scanned.term, position.column)
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
