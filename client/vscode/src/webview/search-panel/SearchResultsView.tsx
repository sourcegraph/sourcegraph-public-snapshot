import classNames from 'classnames'
import React, { useCallback, useEffect, useMemo, useState } from 'react'

import {
    SearchPatternType,
    fetchAutoDefinedSearchContexts,
    getUserSearchContextNamespaces,
    fetchSearchContexts,
} from '@sourcegraph/search'
import { SearchBox, StreamingSearchResultsList } from '@sourcegraph/search-ui'
import { fetchHighlightedFileLineRanges } from '@sourcegraph/shared/src/backend/file'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { LATEST_VERSION } from '@sourcegraph/shared/src/search/stream'
import { globbingEnabledFromSettings } from '@sourcegraph/shared/src/util/globbing'

import { SearchResultsState } from '../../state'
import { WebviewPageProps } from '../platform/context'

// TODO:
// SearchBox (ensure "manage contexts" button works)
// Search results infobar (maybe in forked_components folder?)
// Sign up CTA (for unauthenticated users)
// DidYouMean (for suggestions. move to search-ui)
// StreamingSearchResultsList

export interface SearchResultsViewProps extends WebviewPageProps {
    context: SearchResultsState['context']
}

export const SearchResultsView: React.FunctionComponent<SearchResultsViewProps> = ({
    extensionCoreAPI,
    authenticatedUser,
    platformContext,
    settingsCascade,
    theme,
    context,
}) => {
    // Toggling case sensitivity or pattern type should trigger a new search for results view.

    const [userQueryState, setUserQueryState] = useState(context.submittedSearchQueryState.queryState)

    // Update local query state on e.g. sidebar events.
    useEffect(() => {
        setUserQueryState(context.submittedSearchQueryState.queryState)
    }, [context.submittedSearchQueryState.queryState])

    const onSubmit = useCallback(
        (options?: { caseSensitive?: boolean; patternType?: SearchPatternType }) => {
            const previousSearchQueryState = context.submittedSearchQueryState

            extensionCoreAPI
                .streamSearch(userQueryState.query, {
                    caseSensitive: options?.caseSensitive ?? previousSearchQueryState.searchCaseSensitivity,
                    patternType: options?.patternType ?? previousSearchQueryState.searchPatternType,
                    version: LATEST_VERSION,
                    trace: undefined,
                })
                .then(() => {})
                .catch(() => {})
        },
        [userQueryState.query, context.submittedSearchQueryState, extensionCoreAPI]
    )

    // Submit new search
    const setCaseSensitivity = useCallback(
        (caseSensitive: boolean) => {
            onSubmit({ caseSensitive })
        },
        [onSubmit]
    )

    const setPatternType = useCallback(
        (patternType: SearchPatternType) => {
            onSubmit({ patternType })
        },
        [onSubmit]
    )

    const fetchHighlightedFileLineRangesWithContext = useCallback(
        (parameters: FetchFileParameters) => fetchHighlightedFileLineRanges({ ...parameters, platformContext }),
        [platformContext]
    )

    const globbing = useMemo(() => globbingEnabledFromSettings(settingsCascade), [settingsCascade])

    return (
        <div>
            <div className="d-flex my-2">
                <SearchBox
                    caseSensitive={context.submittedSearchQueryState?.searchCaseSensitivity}
                    setCaseSensitivity={setCaseSensitivity}
                    patternType={SearchPatternType.literal}
                    setPatternType={setPatternType}
                    isSourcegraphDotCom={true}
                    hasUserAddedExternalServices={false}
                    hasUserAddedRepositories={true} // Used for search context CTA, which we won't show here.
                    structuralSearchDisabled={false}
                    queryState={userQueryState}
                    onChange={setUserQueryState}
                    onSubmit={onSubmit}
                    authenticatedUser={authenticatedUser}
                    searchContextsEnabled={true}
                    showSearchContext={true}
                    showSearchContextManagement={false} // Enable this after refactoring
                    defaultSearchContextSpec="global"
                    setSelectedSearchContextSpec={() => {}} // TODO state machine emit
                    selectedSearchContextSpec="global"
                    fetchSearchContexts={fetchSearchContexts}
                    fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                    getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                    settingsCascade={settingsCascade}
                    globbing={globbing}
                    isLightTheme={theme === 'theme-light'}
                    telemetryService={platformContext.telemetryService}
                    platformContext={platformContext}
                    // TODO editor font
                    className={classNames('flex-grow-1 flex-shrink-past-contents')}
                />
            </div>
            {/* <SearchResultsInfobar /> */}
            {/* <SignUpCta /> */}
            {/* <DidYouMean /> */}
            <StreamingSearchResultsList
                isLightTheme={theme === 'theme-light'}
                settingsCascade={settingsCascade}
                telemetryService={platformContext.telemetryService}
                allExpanded={false}
                isSourcegraphDotCom={false} // TODO
                searchContextsEnabled={true}
                showSearchContext={true}
                platformContext={platformContext}
                results={context.searchResults ?? undefined}
                authenticatedUser={authenticatedUser}
                fetchHighlightedFileLineRanges={fetchHighlightedFileLineRangesWithContext}
                executedQuery={context.submittedSearchQueryState.queryState.query}
            />
        </div>
    )
}
