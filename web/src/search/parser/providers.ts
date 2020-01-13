import * as Monaco from 'monaco-editor'
import { Observable, fromEventPattern, of } from 'rxjs'
import { parseSearchQuery } from './parser'
import { map, first, takeUntil, publishReplay, refCount, switchMap } from 'rxjs/operators'
import { getMonacoTokens } from './tokens'
import { getDiagnostics } from './diagnostics'
import { getCompletionItems } from './completion'
import { SearchSuggestion } from '../../../../shared/src/graphql/schema'
import { getHoverResult } from './hover'

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

/**
 * Returns the providers used by the Monaco query input to provide syntax highlighting,
 * hovers, completions and diagnostics for the Sourcegraph search syntax.
 */
export function getProviders(
    searchQueries: Observable<string>,
    fetchSuggestions: (input: string) => Observable<SearchSuggestion[]>
): SearchFieldProviders {
    const parsedQueries = searchQueries.pipe(
        map(rawQuery => {
            const parsed = parseSearchQuery(rawQuery)
            return { rawQuery, parsed }
        }),
        publishReplay(1),
        refCount()
    )
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
            provideHover: (_, position, token) =>
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
            triggerCharacters: [
                ':',
                'a',
                'b',
                'c',
                'd',
                'e',
                'f',
                'g',
                'h',
                'i',
                'j',
                'k',
                'l',
                'm',
                'n',
                'o',
                'p',
                'q',
                'r',
                's',
                't',
                'u',
                'v',
                'w',
                'x',
                'y',
                'z',
                '-',
            ],
            provideCompletionItems: (_, position, context, token) =>
                parsedQueries
                    .pipe(
                        first(),
                        switchMap(({ rawQuery, parsed }) =>
                            parsed.type === 'error'
                                ? of(null)
                                : getCompletionItems(rawQuery, parsed.token, position, context, fetchSuggestions)
                        ),
                        takeUntil(fromEventPattern(handler => token.onCancellationRequested(handler)))
                    )
                    .toPromise(),
        },
        diagnostics: parsedQueries.pipe(
            map(({ parsed }) => (parsed.type === 'success' ? getDiagnostics(parsed.token) : []))
        ),
    }
}
