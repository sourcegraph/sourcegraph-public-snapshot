import * as H from 'history'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { SearchPatternType } from '../../../../../shared/src/graphql-operations'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { VersionContextProps } from '../../../../../shared/src/search/util'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { AuthenticatedUser } from '../../../auth'
import { PageTitle } from '../../../components/PageTitle'
import { QueryState } from '../../helpers'
import { LATEST_VERSION } from '../SearchResults'
import { SearchResultsInfoBar } from '../SearchResultsInfoBar'
import { SearchResultTypeTabs } from '../SearchResultTypeTabs'
import { StreamingProgress } from './progress/StreamingProgress'
import { StreamingSearchResultsFilterBars } from './StreamingSearchResultsFilterBars'
import { CaseSensitivityProps, parseSearchURL, PatternTypeProps, SearchStreamingProps } from '../..'

export interface StreamingSearchResultsProps
    extends SearchStreamingProps,
        PatternTypeProps,
        VersionContextProps,
        CaseSensitivityProps,
        SettingsCascadeProps,
        ExtensionsControllerProps<'executeCommand' | 'extHostAPI' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    navbarSearchQueryState: QueryState
}

export const StreamingSearchResults: React.FunctionComponent<StreamingSearchResultsProps> = props => {
    const {
        patternType: currentPatternType,
        setPatternType,
        caseSensitive: currentCaseSensitive,
        setCaseSensitivity,
        streamSearch,
    } = props

    const { query = '', patternType, caseSensitive, versionContext } = parseSearchURL(props.location.search)

    useEffect(() => {
        if (patternType && patternType !== currentPatternType) {
            setPatternType(patternType)
        }
    }, [patternType, currentPatternType, setPatternType])

    useEffect(() => {
        if (caseSensitive && caseSensitive !== currentCaseSensitive) {
            setCaseSensitivity(caseSensitive)
        }
    }, [caseSensitive, currentCaseSensitive, setCaseSensitivity])

    const results = useObservable(
        useMemo(
            () =>
                streamSearch(
                    caseSensitive ? `${query} case:yes` : query,
                    LATEST_VERSION,
                    patternType ?? SearchPatternType.literal,
                    versionContext
                ),
            [streamSearch, caseSensitive, query, patternType, versionContext]
        )
    )

    const [allExpanded, setAllExpanded] = useState(false)
    const onExpandAllResultsToggle = useCallback(() => setAllExpanded(oldValue => !oldValue), [setAllExpanded])

    const onDidCreateSavedQuery = useCallback(() => {}, [])
    const onSaveQueryClick = useCallback(() => {}, [])
    const didSave = false

    return (
        <div className="test-search-results search-results d-flex flex-column w-100">
            <PageTitle key="page-title" title={query} />
            <StreamingSearchResultsFilterBars {...props} results={results} />
            <div className="search-results-list">
                <div className="d-lg-flex mb-2 align-items-end flex-wrap">
                    <SearchResultTypeTabs
                        {...props}
                        query={props.navbarSearchQueryState.query}
                        className="search-results-list__tabs"
                    />

                    <SearchResultsInfoBar
                        {...props}
                        query={query}
                        resultsFound={results ? results.results.length > 0 : false}
                        className="border-bottom flex-grow-1"
                        allExpanded={allExpanded}
                        onExpandAllResultsToggle={onExpandAllResultsToggle}
                        onSaveQueryClick={onSaveQueryClick}
                        onDidCreateSavedQuery={onDidCreateSavedQuery}
                        didSave={didSave}
                        stats={<StreamingProgress progress={results?.progress} />}
                    />
                </div>
            </div>
        </div>
    )
}
