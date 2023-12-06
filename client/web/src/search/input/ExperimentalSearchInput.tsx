import { type FC, type PropsWithChildren, useCallback, useEffect, useMemo, useRef } from 'react'

import { Prec } from '@codemirror/state'

// This component makes the experimental search input accessible in the web app
// eslint-disable-next-line no-restricted-imports
import {
    type Action,
    CodeMirrorQueryInputWrapper,
    type CodeMirrorQueryInputWrapperProps,
    type Example,
    exampleSuggestions,
    lastUsedContextSuggestion,
    searchHistoryExtension,
    selectionListener,
} from '@sourcegraph/branded/src/search-ui/experimental'
import type { Editor } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import type { SearchContextProps, SubmitSearchParameters } from '@sourcegraph/shared/src/search'
import { FILTERS, FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { resolveFilterMemoized } from '@sourcegraph/shared/src/search/query/utils'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { createSuggestionsSource, type SuggestionsSourceConfig } from './suggestions'
import { useRecentSearches } from './useRecentSearches'

const examples: Example[] = [
    {
        label: 'repo:has.path()',
        snippet: 'repo:has.path(${}) ${}',
        description: 'Search in repositories containing a path',
        valid: tokens => !tokens.some(token => token.type === 'filter' && token.value?.value.startsWith('has.path(')),
    },
    {
        label: 'repo:has.content()',
        snippet: 'repo:has.content(${}) ${}',
        description: 'Search in repositories with files having specific contents',
        valid: tokens =>
            !tokens.some(token => token.type === 'filter' && token.value?.value.startsWith('has.content(')),
    },
    {
        label: '-file:',
        description: FILTERS[FilterType.file].description(true),
        valid: tokens => !tokens.some(token => token.type === 'filter' && token.field.value === '-file'),
    },
    {
        label: '-repo:',
        description: FILTERS[FilterType.repo].description(true),
        valid: tokens => !tokens.some(token => token.type === 'filter' && token.field.value === '-repo'),
    },
    {
        label: 'repo:my-org.*/.*-cli$',
        // eslint-disable-next-line no-template-curly-in-string
        snippet: 'repo:${my-org.*/.*-cli$} ${}',
        description: 'Search in repositories matching a pattern',
        valid: tokens =>
            !tokens.some(
                token => token.type === 'filter' && resolveFilterMemoized(token.field.value)?.type === FilterType.repo
            ),
    },
    {
        label: 'type:diff select:commit.diff.removed TODO',
        // eslint-disable-next-line no-template-curly-in-string
        snippet: 'type:diff select:commit.diff.removed repo:${my-repo} TODO ${}',
        description: 'Find commits that removed "TODO"',
        valid: tokens => !tokens.some(token => token.type === 'filter' && token.value?.value.startsWith('commit.diff')),
    },
]

function useUsedExamples(): [Set<string>, (value: string) => void] {
    const [usedExamples = [], setUsedExamples] = useTemporarySetting('search.input.usedExamples', [])

    const addUsedExample = useCallback(
        (example: string) => {
            setUsedExamples(examples => (!examples || examples.includes(example) ? examples : [...examples, example]))
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
        TelemetryV2Props,
        Omit<SuggestionsSourceConfig, 'getSearchContext'> {
    submitSearch(parameters: Partial<SubmitSearchParameters>): void
}

/**
 * Experimental search input component. Provides query and history suggestions.
 */
export const ExperimentalSearchInput: FC<PropsWithChildren<ExperimentalSearchInputProps>> = ({
    children,
    telemetryService,
    telemetryRecorder,
    platformContext,
    authenticatedUser,
    fetchSearchContexts,
    getUserSearchContextNamespaces,
    isSourcegraphDotCom,
    submitSearch,
    selectedSearchContextSpec,
    visualMode,
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
                telemetryRecorder.recordEvent(`SearchInput${eventNameMap[action.type]}`, 'rendered', {
                    privateMetadata: { type: option.kind, source },
                })
            }),
            Prec.low(
                exampleSuggestions({
                    getUsedExamples: () => usedExamplesRef.current,
                    markExampleUsed: addExample,
                    examples,
                })
            ),
        ],
        [telemetryService, telemetryRecorder, addExample]
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
            visualMode={visualMode}
            className={inputProps.className}
        >
            {children}
        </CodeMirrorQueryInputWrapper>
    )
}
