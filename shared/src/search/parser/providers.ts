import * as Monaco from 'monaco-editor'
import { Observable, fromEventPattern, of, combineLatest } from 'rxjs'
import { parseSearchQuery } from './parser'
import { map, first, takeUntil, publishReplay, refCount, switchMap, debounceTime, share } from 'rxjs/operators'
import { getMonacoTokens } from './tokens'
import { getDiagnostics } from './diagnostics'
import { getCompletionItems } from './completion'
import { getHoverResult } from './hover'
import { SearchPatternType } from '../../graphql/schema'
import { SearchSuggestion } from '../suggestions'

interface SearchFieldProviders {
    tokens: Monaco.languages.TokensProvider
    hover: Monaco.languages.HoverProvider
    completion: Monaco.languages.CompletionItemProvider
    diagnostics: Observable<Monaco.editor.IMarkerData[]>
}

/**
 * A dummy parsing state, required for the token provider.
 */
const PARSER_STATE: Monaco.languages.IState = {
    clone: () => ({ ...PARSER_STATE }),
    equals: () => false,
}

const alphabet = 'abcdefghijklmnopqrstuvwxyz'

/**
 * Returns the providers used by the Monaco query input to provide syntax highlighting,
 * hovers, completions and diagnostics for the Sourcegraph search syntax.
 */
export function getProviders(
    searchQueries: Observable<string>,
    patternTypes: Observable<SearchPatternType>,
    fetchSuggestions: (input: string) => Observable<SearchSuggestion[]>,
    globbing: boolean
): SearchFieldProviders {
    const parsedQueries = searchQueries.pipe(
        map(rawQuery => {
            const parsed = parseSearchQuery(rawQuery)
            return { rawQuery, parsed }
        }),
        publishReplay(1),
        refCount()
    )
    const debouncedDynamicSuggestions = searchQueries.pipe(debounceTime(300), switchMap(fetchSuggestions), share())
    return {
        tokens: {
            getInitialState: () => PARSER_STATE,
            tokenize: line => {
                const result = parseSearchQuery(line)
                if (result.type === 'success') {
                    return {
                        tokens: getMonacoTokens(result.token),
                        endState: PARSER_STATE,
                    }
                }
                return { endState: PARSER_STATE, tokens: [] }
            },
        },
        hover: {
            provideHover: (textModel, position, token) =>
                parsedQueries
                    .pipe(
                        first(),
                        map(({ parsed }) => (parsed.type === 'error' ? null : getHoverResult(parsed.token, position))),
                        takeUntil(fromEventPattern(handler => token.onCancellationRequested(handler)))
                    )
                    .toPromise(),
        },
        completion: {
            // An explicit list of trigger characters is needed for the Monaco editor to show completions.
            triggerCharacters: [':', '-', ...alphabet, ...alphabet.toUpperCase()],
            provideCompletionItems: (textModel, position, context, token) =>
                parsedQueries
                    .pipe(
                        first(),
                        switchMap(({ parsed }) =>
                            parsed.type === 'error'
                                ? of(null)
                                : getCompletionItems(parsed.token, position, debouncedDynamicSuggestions, globbing)
                        ),
                        takeUntil(fromEventPattern(handler => token.onCancellationRequested(handler)))
                    )
                    .toPromise(),
        },
        diagnostics: combineLatest([parsedQueries, patternTypes]).pipe(
            map(([{ parsed }, patternType]) =>
                parsed.type === 'success' ? getDiagnostics(parsed.token, patternType) : []
            )
        ),
    }
}
