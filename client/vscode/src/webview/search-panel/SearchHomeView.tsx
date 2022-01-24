import classNames from 'classnames'
import React, { useCallback, useMemo, useState } from 'react'

import {
    SearchPatternType,
    fetchAutoDefinedSearchContexts,
    getUserSearchContextNamespaces,
    fetchSearchContexts,
} from '@sourcegraph/search'
import { SearchBox } from '@sourcegraph/search-ui'
import { LATEST_VERSION } from '@sourcegraph/shared/src/search/stream'
import { globbingEnabledFromSettings } from '@sourcegraph/shared/src/util/globbing'

import { SearchHomeState } from '../../state'
import { WebviewPageProps } from '../platform/context'

// TODO:
// Logo, feedback button
// SearchBox
// Search examples

export interface SearchHomeViewProps extends WebviewPageProps {
    context: SearchHomeState['context']
}

export const SearchHomeView: React.FunctionComponent<SearchHomeViewProps> = ({
    extensionCoreAPI,
    authenticatedUser,
    platformContext,
    settingsCascade,
    theme,
}) => {
    // Toggling case sensitivity or pattern type does NOT trigger a new search on home view.
    const [caseSensitive, setCaseSensitivity] = useState(false)
    const [patternType, setPatternType] = useState(SearchPatternType.literal)

    const [userQueryState, setUserQueryState] = useState({
        query: '',
    })

    // TODO: get query state from sidebar in context (already implemented "for free" w/ submission for search results view).

    const onSubmit = useCallback(() => {
        // TODO check if we need query state ref for perf
        extensionCoreAPI
            .streamSearch(userQueryState.query, {
                caseSensitive,
                patternType,
                version: LATEST_VERSION,
                trace: undefined,
            })
            .then(() => {})
            .catch(() => {})
    }, [userQueryState.query, caseSensitive, patternType, extensionCoreAPI])

    const globbing = useMemo(() => globbingEnabledFromSettings(settingsCascade), [settingsCascade])

    return (
        <div>
            {/* <BrandHeader /> */}
            <div className="d-flex my-2">
                <SearchBox
                    caseSensitive={caseSensitive}
                    setCaseSensitivity={setCaseSensitivity}
                    patternType={patternType}
                    setPatternType={setPatternType}
                    isSourcegraphDotCom={false} // TODO
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
            {/* <SearchExamples /> */}
        </div>
    )
}
