import * as Monaco from 'monaco-editor'
import { Observable, fromEventPattern, of } from 'rxjs'
import { map, delay, takeUntil, switchMap } from 'rxjs/operators'

import { SearchPatternType } from '../../graphql-operations'
import { isSearchMatchOfType, SearchMatch } from '../stream'

import { getCompletionItems, REPO_DEPS_PREDICATE_REGEX } from './completion'
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

/**
 * getSuggestionsQuery might return an empty query. The caller is responsible
 * for handling this accordingly.
 */
export function getSuggestionQuery(tokens: Token[], tokenAtColumn: Token, suggestionType: SearchMatch['type']): string {
    const hasAndOrOperators = tokens.some(
        token => token.type === 'keyword' && (token.kind === KeywordKind.Or || token.kind === KeywordKind.And)
    )

    let tokenValue = ''

    switch (tokenAtColumn.type) {
        case 'filter':
            tokenValue = tokenAtColumn.value?.value ?? ''
            break
        case 'pattern':
            tokenValue = tokenAtColumn.value
            break
    }

    if (!tokenValue) {
        return ''
    }

    if (suggestionType === 'repo') {
        const depsPredicateMatch = tokenValue.match(REPO_DEPS_PREDICATE_REGEX)
        const repoValue = depsPredicateMatch ? depsPredicateMatch[2] : tokenValue
        const relevantFilters =
            !hasAndOrOperators && !depsPredicateMatch ? serializeFilters(tokens, REPO_SUGGESTION_FILTERS) : ''
        return `${relevantFilters} repo:${repoValue} type:repo count:${MAX_SUGGESTION_COUNT}`.trimStart()
    }

    // For the cases below, we are not handling queries with and/or operators. This is because we would need to figure out
    // for each filter which filters from the surrounding expression apply to it. For example, if we have a query: `repo:x file:y z OR repo:xx file:yy`
    // and we want to get suggestions for the `file:yy` filter. We would only want to include file suggestions from the `xx` repo and not the `x` repo, because it
    // is a part of a different expression.
    if (hasAndOrOperators) {
        return ''
    }

    if (suggestionType === 'path') {
        const relevantFilters = serializeFilters(tokens, FILE_SUGGESTION_FILTERS)
        return `${relevantFilters} file:${tokenValue} type:path count:${MAX_SUGGESTION_COUNT}`.trimStart()
    }
    if (suggestionType === 'symbol') {
        const relevantFilters = serializeFilters(tokens, [...FILE_SUGGESTION_FILTERS, FilterType.file])
        return `${relevantFilters} ${tokenValue} type:symbol count:${MAX_SUGGESTION_COUNT}`.trimStart()
    }

    return ''
}

function getTokenAtColumn(tokens: Token[], column: number): Token | null {
    return tokens.find(({ range }) => range.start + 1 <= column && range.end + 1 >= column) ?? null
}

export function createCancelableFetchSuggestions(
    fetchSuggestions: (query: string) => Observable<SearchMatch[]>
): (query: string, onAbort: (hander: () => void) => void) => Promise<SearchMatch[]> {
    return (query, onAbort) => {
        if (!query) {
            // Don't fetch suggestions if the query is empty. This would result
            // in arbitrary result types being returned, which is unexpected.
            return Promise.resolve([])
        }

        let aborted = false

        // By listeing to the abort event of the autocompletion we
        // can close the connection to server early and don't have to download
        // data sent by the server.
        const abort = new Observable(subscriber => {
            onAbort(() => {
                aborted = true
                subscriber.next(null)
                subscriber.complete()
            })
        })

        return (
            of(query)
                .pipe(
                    // We use a delay here to implement a custom debounce. In the
                    // next step we check if the current completion request was
                    // cancelled in the meantime.
                    // This prevents us from needlessly running multiple suggestion
                    // queries.
                    delay(200),
                    switchMap(query => (aborted ? Promise.resolve([]) : fetchSuggestions(query))),
                    takeUntil(abort)
                )
                // toPromise may return undefined if the observable completes before
                // a value was emitted . The return type was fixed in newer versions
                // (and the method was actually deprecated).
                // See https://rxjs.dev/deprecations/to-promise
                .toPromise()
                .then(result => result ?? [])
        )
    }
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
