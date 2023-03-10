import { FC, PropsWithChildren, useEffect, useMemo, useRef } from 'react'

// The experimental search input should be shown in the navbar
// eslint-disable-next-line no-restricted-imports
import {
    Action,
    CodeMirrorQueryInputWrapper,
    CodeMirrorQueryInputWrapperProps,
    searchHistoryExtension,
    selectionListener,
} from '@sourcegraph/branded/src/search-ui/experimental'
import { SubmitSearchParameters } from '@sourcegraph/shared/src/search'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { createSuggestionsSource, SuggestionsSourceConfig } from './suggestions'
import { useRecentSearches } from './useRecentSearches'

const eventNameMap: Record<Action['type'], string> = {
    completion: 'Add',
    goto: 'GoTo',
    command: 'Command',
}

export interface ExperimentalSearchInputProps
    extends Omit<CodeMirrorQueryInputWrapperProps, 'suggestionSource' | 'extensions' | 'placeholder'>,
        TelemetryProps,
        SuggestionsSourceConfig {
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
    ...inputProps
}) => {
    const { recentSearches } = useRecentSearches()
    const recentSearchesRef = useRef(recentSearches)
    useEffect(() => {
        recentSearchesRef.current = recentSearches
    }, [recentSearches])

    const submitSearchRef = useRef(submitSearch)
    useEffect(() => {
        submitSearchRef.current = submitSearch
    }, [submitSearch])

    const suggestionSource = useMemo(
        () =>
            createSuggestionsSource({
                platformContext,
                authenticatedUser,
                fetchSearchContexts,
                getUserSearchContextNamespaces,
                isSourcegraphDotCom,
            }),
        [platformContext, authenticatedUser, fetchSearchContexts, getUserSearchContextNamespaces, isSourcegraphDotCom]
    )

    const extensions = useMemo(
        () => [
            searchHistoryExtension({
                mode: {
                    name: 'History',
                    placeholder: 'Filter history',
                },
                source: () => recentSearchesRef.current ?? [],
                submitQuery: query => submitSearchRef.current?.({ query }),
            }),
            selectionListener.of(({ option, action, source }) => {
                telemetryService.log(`SearchInput${eventNameMap[action.type]}`, {
                    type: option.kind,
                    source,
                })
            }),
        ],
        [telemetryService]
    )

    return (
        <CodeMirrorQueryInputWrapper
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
