import { useEffect, useMemo } from 'react'

import * as Monaco from 'monaco-editor'
import { Observable } from 'rxjs'
import * as uuid from 'uuid'

import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { getDiagnostics } from '@sourcegraph/shared/src/search/query/diagnostics'
import { getProviders } from '@sourcegraph/shared/src/search/query/providers'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { SearchMatch } from '@sourcegraph/shared/src/search/stream'

const SOURCEGRAPH_SEARCH = 'sourcegraphSearch' as const

/**
 * Adds code intelligence for the Sourcegraph search syntax to Monaco.
 */
export function useQueryIntelligence(
    fetchSuggestions: (query: string) => Observable<SearchMatch[]>,
    options: {
        patternType: SearchPatternType
        globbing: boolean
        interpretComments?: boolean
        isSourcegraphDotCom?: boolean
        disablePatternSuggestions?: boolean
    }
): string {
    // Due to the global nature of Monaco (tokens, hover, completion) providers we have to create a unique
    // language for each editor and register the providers for the new language. This ensures that there is no cross-contamination
    // between different editors using the query intelligence hook. The main issue with using a single language id
    // is when a component using the hook gets unmounted. When navigating between pages that both contain a search box (homepage -> search results page),
    // the unmounted useEffect hook below would trigger cleanup after the search results hook already registered its providers. This effectively
    // removes query intelligence from all editors.
    const sourcegraphSearchLanguageId = useMemo(() => `${SOURCEGRAPH_SEARCH}-${uuid.v4()}`, [])

    const memoizedOptions = useMemo(
        () => ({
            patternType: options.patternType,
            globbing: options.globbing,
            interpretComments: options.interpretComments,
            isSourcegraphDotCom: options.isSourcegraphDotCom,
            disablePatternSuggestions: options.disablePatternSuggestions,
        }),
        [
            options.patternType,
            options.globbing,
            options.interpretComments,
            options.isSourcegraphDotCom,
            options.disablePatternSuggestions,
        ]
    )

    useEffect(() => {
        if (Monaco.languages.getEncodedLanguageId(sourcegraphSearchLanguageId) === 0) {
            // Register language ID
            Monaco.languages.register({ id: sourcegraphSearchLanguageId })
        }

        // Register providers
        const providers = getProviders(fetchSuggestions, memoizedOptions)
        const disposables = [
            Monaco.languages.setTokensProvider(sourcegraphSearchLanguageId, providers.tokens),
            Monaco.languages.registerHoverProvider(sourcegraphSearchLanguageId, providers.hover),
            Monaco.languages.registerCompletionItemProvider(sourcegraphSearchLanguageId, providers.completion),
        ]
        return () => {
            for (const disposable of disposables) {
                disposable.dispose()
            }
        }
    }, [fetchSuggestions, sourcegraphSearchLanguageId, memoizedOptions])

    return sourcegraphSearchLanguageId
}

/**
 * Adds diagnostic markers for the Sourcegraph search syntax to Monaco.
 * For example, it adds a marker if the query contains a filter with an invalid value: `patterntype:invalid`.
 */
export function useQueryDiagnostics(
    editor: Monaco.editor.IStandaloneCodeEditor | undefined,
    options: {
        patternType: SearchPatternType
        interpretComments?: boolean
    }
): void {
    const memoizedOptions = useMemo(
        () => ({
            patternType: options.patternType,
            interpretComments: options.interpretComments,
        }),
        [options.patternType, options.interpretComments]
    )

    useEffect(() => {
        if (!editor) {
            return
        }
        const disposable = editor.onDidChangeModelContent(() => {
            const model = editor.getModel()
            if (!model) {
                return
            }
            const scanned = scanSearchQuery(
                model.getValue(),
                memoizedOptions.interpretComments ?? false,
                memoizedOptions.patternType
            )
            const markers = scanned.type === 'success' ? getDiagnostics(scanned.term, memoizedOptions.patternType) : []
            Monaco.editor.setModelMarkers(model, 'diagnostics', markers)
        })
        return () => disposable.dispose()
    }, [editor, memoizedOptions])
}
