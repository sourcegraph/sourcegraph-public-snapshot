import * as Monaco from 'monaco-editor'
import { useEffect } from 'react'
import { Observable } from 'rxjs'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql/schema'
import { getDiagnostics } from '@sourcegraph/shared/src/search/query/diagnostics'
import { getProviders } from '@sourcegraph/shared/src/search/query/providers'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { SearchSuggestion } from '@sourcegraph/shared/src/search/suggestions'

export const SOURCEGRAPH_SEARCH = 'sourcegraphSearch' as const

/**
 * Adds code intelligence for the Sourcegraph search syntax to Monaco.
 */
export function useQueryIntelligence(
    fetchSuggestions: (query: string) => Observable<SearchSuggestion[]>,
    options: {
        patternType: SearchPatternType
        globbing: boolean
        interpretComments?: boolean
        isSourcegraphDotCom?: boolean
    }
): void {
    useEffect(() => {
        // Register language ID
        Monaco.languages.register({ id: SOURCEGRAPH_SEARCH })

        // Register providers
        const providers = getProviders(fetchSuggestions, options)
        const setTokensProviderDisposable = Monaco.languages.setTokensProvider(SOURCEGRAPH_SEARCH, providers.tokens)
        const registerHoverProviderDisposable = Monaco.languages.registerHoverProvider(
            SOURCEGRAPH_SEARCH,
            providers.hover
        )
        const registerCompletionItemProviderDisposable = Monaco.languages.registerCompletionItemProvider(
            SOURCEGRAPH_SEARCH,
            providers.completion
        )

        return () => {
            setTokensProviderDisposable.dispose()
            registerHoverProviderDisposable.dispose()
            registerCompletionItemProviderDisposable.dispose()
        }
    }, [fetchSuggestions, options])
}

export function useQueryDiagnostics(
    editor: Monaco.editor.IStandaloneCodeEditor | undefined,
    options: {
        patternType: SearchPatternType
        interpretComments?: boolean
    }
): void {
    useEffect(() => {
        if (!editor) {
            return
        }
        const disposable = editor.onDidChangeModelContent(() => {
            const model = editor.getModel()
            if (!model) {
                return
            }
            const scanned = scanSearchQuery(model.getValue(), options.interpretComments ?? false, options.patternType)
            const markers = scanned.type === 'success' ? getDiagnostics(scanned.term, options.patternType) : []
            Monaco.editor.setModelMarkers(model, 'diagnostics', markers)
        })
        return () => disposable.dispose()
    }, [editor, options])
}
