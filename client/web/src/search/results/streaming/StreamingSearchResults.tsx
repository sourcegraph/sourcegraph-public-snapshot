import * as H from 'history'
import React, { useCallback, useEffect, useMemo } from 'react'
import { CaseSensitivityProps, parseSearchURL, PatternTypeProps, SearchStreamingProps } from '../..'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { SearchPatternType } from '../../../../../shared/src/graphql-operations'
import { VersionContextProps } from '../../../../../shared/src/search/util'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { PageTitle } from '../../../components/PageTitle'
import { Settings } from '../../../schema/settings.schema'
import { submitSearch, toggleSearchFilter, QueryState } from '../../helpers'
import { Filter } from '../../stream'
import { LATEST_VERSION } from '../SearchResults'
import { DynamicSearchFilter, SearchResultsFilterBars } from '../SearchResultsFilterBars'
import {
    isSettingsValid,
    SettingsCascadeOrError,
    SettingsCascadeProps,
} from '../../../../../shared/src/settings/settings'

interface StreamingSearchResultsProps
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
        settingsCascade,
        extensionsController,
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

    const contributions = useObservable(
        useMemo(() => extensionsController.services.contribution.getContributions(), [extensionsController])
    )
    const filters = results?.filters
    const quickLinks = (isSettingsValid<Settings>(settingsCascade) && settingsCascade.final.quicklinks) || []

    const onDynamicFilterClicked = useCallback(
        (value: string) => {
            props.telemetryService.log('DynamicFilterClicked', {
                search_filter: { value },
            })

            const newQuery = toggleSearchFilter(props.navbarSearchQueryState.query, value)

            submitSearch({ ...props, query: newQuery, source: 'filter' })
        },
        [props]
    )
    const showMoreResults = useCallback(() => {}, [])
    const calculateCount = useCallback(() => 0, [])

    return (
        <div className="test-search-results search-results d-flex flex-column w-100">
            <PageTitle key="page-title" title={query} />
            <SearchResultsFilterBars
                navbarSearchQuery={props.navbarSearchQueryState.query}
                searchSucceeded={!!results}
                resultsLimitHit={
                    !!results && results.progress.skipped.some(skipped => skipped.reason.includes('-limit'))
                }
                genericFilters={getFilters(filters, settingsCascade)}
                extensionFilters={contributions?.searchFilters}
                repoFilters={getRepoFilters(filters)}
                quickLinks={quickLinks}
                onFilterClick={onDynamicFilterClicked}
                onShowMoreResultsClick={showMoreResults}
                calculateShowMoreResultsCount={calculateCount}
            />
        </div>
    )
}

/** Combines dynamic filters and search scopes into a list de-duplicated by value. */
function getFilters(
    resultFilters: Filter[] | undefined,
    settingsCascade: SettingsCascadeOrError<Settings>
): DynamicSearchFilter[] {
    const filters = new Map<string, DynamicSearchFilter>()

    if (resultFilters) {
        const dynamicFilters = resultFilters.filter(filter => filter.kind !== 'repo')
        for (const filter of dynamicFilters) {
            filters.set(filter.value, filter)
        }
    }

    const scopes = (isSettingsValid<Settings>(settingsCascade) && settingsCascade.final['search.scopes']) || []
    if (resultFilters) {
        for (const scope of scopes) {
            if (!filters.has(scope.value)) {
                filters.set(scope.value, scope)
            }
        }
    } else {
        for (const scope of scopes) {
            // Check for if filter.value already exists and if so, overwrite with user's configured scope name.
            const existingFilter = filters.get(scope.value)
            // This works because user setting configs are the last to be processed after Global and Org.
            // Thus, user set filters overwrite the equal valued existing filters.
            if (existingFilter) {
                existingFilter.name = scope.name || scope.value
            }
            filters.set(scope.value, existingFilter || scope)
        }
    }

    return [...filters.values()]
}

function getRepoFilters(resultFilters: Filter[] | undefined): DynamicSearchFilter[] | undefined {
    if (resultFilters) {
        return resultFilters
            .filter(filter => filter.kind === 'repo' && filter.value !== '')
            .map(filter => ({
                name: filter.label,
                value: filter.value,
                count: filter.count,
                limitHit: filter.limitHit,
            }))
    }
    return undefined
}
