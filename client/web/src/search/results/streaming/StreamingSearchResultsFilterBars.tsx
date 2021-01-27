import * as H from 'history'
import React, { useCallback, useMemo } from 'react'
import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '../..'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { VersionContextProps } from '../../../../../shared/src/search/util'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { Settings } from '../../../schema/settings.schema'
import { QueryState, submitSearch, toggleSearchFilter } from '../../helpers'
import { AggregateStreamingSearchResults, Filter } from '../../stream'
import { DynamicSearchFilter, SearchResultsFilterBars } from '../SearchResultsFilterBars'
import {
    isSettingsValid,
    SettingsCascadeOrError,
    SettingsCascadeProps,
} from '../../../../../shared/src/settings/settings'
import { from } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { wrapRemoteObservable } from '../../../../../shared/src/api/client/api/common'

interface Props
    extends SettingsCascadeProps,
        ExtensionsControllerProps<'executeCommand' | 'extHostAPI' | 'services'>,
        TelemetryProps,
        Pick<PatternTypeProps, 'patternType'>,
        Pick<VersionContextProps, 'versionContext'>,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    location: H.Location
    history: H.History

    navbarSearchQueryState: QueryState

    results?: AggregateStreamingSearchResults
}

export const StreamingSearchResultsFilterBars: React.FunctionComponent<Props> = props => {
    const { extensionsController, results, settingsCascade } = props

    const contributions = useObservable(
        useMemo(
            () =>
                from(extensionsController.extHostAPI).pipe(
                    switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getContributions()))
                ),
            [extensionsController]
        )
    )

    const filters = props.results?.filters
    const quickLinks = (isSettingsValid<Settings>(settingsCascade) && settingsCascade.final.quicklinks) || []

    const genericFilters = useMemo(() => getFilters(filters, settingsCascade), [filters, settingsCascade])
    const repoFilters = useMemo(() => getRepoFilters(filters), [filters])

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
        <SearchResultsFilterBars
            navbarSearchQuery={props.navbarSearchQueryState.query}
            searchSucceeded={!!results}
            resultsLimitHit={!!results && results.progress.skipped.some(skipped => skipped.reason.includes('-limit'))}
            genericFilters={genericFilters}
            extensionFilters={contributions?.searchFilters}
            repoFilters={repoFilters}
            quickLinks={quickLinks}
            onFilterClick={onDynamicFilterClicked}
            onShowMoreResultsClick={showMoreResults}
            calculateShowMoreResultsCount={calculateCount}
        />
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
