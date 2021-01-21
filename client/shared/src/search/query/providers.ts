import * as Monaco from 'monaco-editor'
import { Observable, ReplaySubject } from 'rxjs'
import { scanSearchQuery } from './scanner'
import { getMonacoTokens } from './decoratedToken'
import { getCompletionItems } from './completion'
import { getHoverResult } from './hover'
import { SearchSuggestion } from '../suggestions'
import { SearchPatternType } from '../../graphql-operations'
import { debounceTime, share, switchMap } from 'rxjs/operators'

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
 * hovers, completions and diagnostics for the Sourcegraph search syntax.
 */
export function getProviders(
    queryRef: {
        current:
            | undefined
            | {
                  rawQuery: string
                  scanned: ReturnType<typeof scanSearchQuery>
              }
    },
    fetchSuggestions: (input: string) => Observable<SearchSuggestion[]>,
    options: {
        patternType: SearchPatternType
        globbing: boolean
        enableSmartQuery: boolean
        interpretComments?: boolean
    }
): SearchFieldProviders {
    const fetchSuggestionsRequests = new ReplaySubject<string>(1)
    const debouncedDynamicSuggestions = fetchSuggestionsRequests.pipe(
        debounceTime(300),
        switchMap(fetchSuggestions),
        share()
    )
    return {
        tokens: {
            getInitialState: () => SCANNER_STATE,
            tokenize: line => {
                const result = scanSearchQuery(line, options.interpretComments ?? false, options.patternType)
                return result.type === 'success'
                    ? {
                          tokens: getMonacoTokens(result.term, options.enableSmartQuery),
                          endState: SCANNER_STATE,
                      }
                    : { endState: SCANNER_STATE, tokens: [] }
            },
        },
        hover: {
            provideHover: (textModel, position, token) => {
                if (!queryRef.current || queryRef.current.scanned.type === 'error') {
                    return null
                }
                return getHoverResult(queryRef.current.scanned.term, position, options.enableSmartQuery)
            },
        },
        completion: {
            // An explicit list of trigger characters is needed for the Monaco editor to show completions.
            triggerCharacters: [...printable, ...latin1Alpha],
            provideCompletionItems: (textModel, position) => {
                if (!queryRef.current || queryRef.current.scanned.type === 'error') {
                    return null
                }
                fetchSuggestionsRequests.next(queryRef.current.rawQuery)
                return getCompletionItems(
                    queryRef.current.scanned.term,
                    position,
                    debouncedDynamicSuggestions,
                    options.globbing
                )
            },
        },
    }
}
