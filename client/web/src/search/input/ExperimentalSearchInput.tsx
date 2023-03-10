import { FC, PropsWithChildren, useCallback, useEffect, useMemo, useRef } from 'react'

import { Prec } from '@codemirror/state'

// This component makes the experimental search input accessible in the web app
// eslint-disable-next-line no-restricted-imports
import {
    type Action,
    CodeMirrorQueryInputWrapper,
    type CodeMirrorQueryInputWrapperProps,
    exampleSuggestions,
    lastUsedContextSuggestion,
    searchHistoryExtension,
    selectionListener,
} from '@sourcegraph/branded/src/search-ui/experimental'
import type { Editor } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import type { SearchContextProps, SubmitSearchParameters } from '@sourcegraph/shared/src/search'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'

import { createSuggestionsSource, type SuggestionsSourceConfig } from './suggestions'
import { useRecentSearches } from './useRecentSearches'

function useUsedExamples(): [Set<string>, (value: string) => void] {
    const [usedExamples = [], setUsedExamples] = useTemporarySetting('search.input.usedExamples', [])
    const usedExamplesRef = useRef(usedExamples)

    useEffect(() => {
        usedExamplesRef.current = usedExamples
    }, [usedExamples])

    const addUsedExample = useCallback(
        (example: string) => {
            if (!usedExamplesRef.current?.includes(example)) {
                setUsedExamples([...usedExamplesRef.current, example])
            }
        },
        [setUsedExamples]
    )

    return [new Set(usedExamples), addUsedExample]
}

const eventNameMap: Record<Action['type'], string> = {
    completion: 'Add',
    goto: 'GoTo',
    command: 'Command',
}

export interface ExperimentalSearchInputProps
    extends Omit<CodeMirrorQueryInputWrapperProps, 'suggestionSource' | 'extensions' | 'placeholder'>,
        Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        TelemetryProps,
        Omit<SuggestionsSourceConfig, 'getSearchContext'> {
    submitSearch(parameters: Partial<SubmitSearchParameters>): void
}

/**
 * Experimental search input component. Provides query and history suggestions.
 */
export const ExperimentalSearchInput: FC<PropsWithChildren<ExperimentalSearchInputProps>> = ({
    children,
    telemetryService,
    platformContext,
    authenticatedUser,
    fetchSearchContexts,
    getUserSearchContextNamespaces,
    isSourcegraphDotCom,
    submitSearch,
    selectedSearchContextSpec,
    ...inputProps
}) => {
    const { recentSearches } = useRecentSearches()
    const recentSearchesRef = useRef(recentSearches)
    useEffect(() => {
        recentSearchesRef.current = recentSearches
    }, [recentSearches])

    const [usedExamples, addExample] = useUsedExamples()
    const usedExamplesRef = useRef(usedExamples)
    useEffect(() => {
        usedExamplesRef.current = usedExamples
    }, [usedExamples])

    const submitSearchRef = useRef(submitSearch)
    useEffect(() => {
        submitSearchRef.current = submitSearch
    }, [submitSearch])

    const getSearchContextRef = useRef(() => selectedSearchContextSpec)
    useEffect(() => {
        getSearchContextRef.current = () => selectedSearchContextSpec
    }, [selectedSearchContextSpec])

    const editorRef = useRef<Editor | null>(null)

    const suggestionSource = useMemo(
        () =>
            createSuggestionsSource({
                platformContext,
                authenticatedUser,
                fetchSearchContexts,
                getUserSearchContextNamespaces,
                isSourcegraphDotCom,
                getSearchContext: () => getSearchContextRef.current(),
            }),
        [platformContext, authenticatedUser, fetchSearchContexts, getUserSearchContextNamespaces, isSourcegraphDotCom]
    )

    const extensions = useMemo(
        () => [
            // Prec ensures that this suggestion is shown first
            Prec.high(lastUsedContextSuggestion({ getContext: () => getSearchContextRef.current() })),
            searchHistoryExtension({
                mode: {
                    name: 'History',
                    placeholder: 'Filter history',
                },
                source: () => recentSearchesRef.current ?? [],
                submitQuery: query => {
                    if (submitSearchRef?.current) {
                        submitSearchRef.current?.({ query })
                        editorRef.current?.blur()
                    }
                },
            }),
            selectionListener.of(({ option, action, source }) => {
                telemetryService.log(`SearchInput${eventNameMap[action.type]}`, {
                    type: option.kind,
                    source,
                })
            }),
            Prec.low(
                exampleSuggestions({
                    getUsedExamples: () => usedExamplesRef.current,
                    markExampleUsed: addExample,
                })
            ),
        ],
        [telemetryService, addExample]
    )

    return (
        <CodeMirrorQueryInputWrapper
            ref={editorRef}
            patternType={inputProps.patternType}
            interpretComments={false}
            queryState={inputProps.queryState}
            onChange={inputProps.onChange}
            onSubmit={inputProps.onSubmit}
            isLightTheme={inputProps.isLightTheme}
            placeholder="Search for code or files..."
            suggestionSource={suggestionSource}
            extensions={extensions}
        >
            {children}
        </CodeMirrorQueryInputWrapper>
    )
}
