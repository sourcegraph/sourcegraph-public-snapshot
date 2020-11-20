import * as H from 'history'
import React, { useEffect, useMemo } from 'react'
import { CaseSensitivityProps, parseSearchURL, PatternTypeProps, SearchStreamingProps } from '../..'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { SearchPatternType } from '../../../../../shared/src/graphql-operations'
import { VersionContextProps } from '../../../../../shared/src/search/util'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { PageTitle } from '../../../components/PageTitle'
import { QueryState } from '../../helpers'
import { LATEST_VERSION } from '../SearchResults'
import { StreamingSearchResultsFilterBars } from './StreamingSearchResultsFilterBars'

export interface StreamingSearchResultsProps
    extends SearchStreamingProps,
        PatternTypeProps,
        VersionContextProps,
        CaseSensitivityProps,
        SettingsCascadeProps,
        ExtensionsControllerProps<'executeCommand' | 'extHostAPI' | 'services'>,
        TelemetryProps {
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
        versionContext,
        streamSearch,
    } = props

    const { query = '', patternType, caseSensitive } = parseSearchURL(props.location.search)

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

    return (
        <div className="test-search-results search-results d-flex flex-column w-100">
            <PageTitle key="page-title" title={query} />
            <StreamingSearchResultsFilterBars {...props} results={results} />
        </div>
    )
}
