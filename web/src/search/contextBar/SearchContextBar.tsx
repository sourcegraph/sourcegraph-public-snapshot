import H from 'history'
import React from 'react'
import { ContributableMenu, SearchFilters } from '../../../../shared/src/api/protocol'
import { RepositoryIcon } from '../../../../shared/src/components/icons'
import { displayRepoName } from '../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { WebActionsNavItems as ActionsNavItems } from '../../components/shared'
import { FilterChip } from '../FilterChip'
import { SearchScopeWithOptionalName } from '../results/SearchResultsFilterBars'

interface Props
    extends ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip'>,
        TelemetryProps {
    results?: Pick<GQL.ISearchResults, 'dynamicFilters'>
    navbarSearchQuery: string
    filters: SearchScopeWithOptionalName[]
    extensionFilters: SearchFilters[] | undefined
    onFilterClick: (value: string) => void

    className?: string
    location: H.Location
}

export const SearchContextBar: React.FunctionComponent<Props> = ({
    results,
    navbarSearchQuery,
    filters,
    extensionFilters,
    onFilterClick,
    className = '',
    ...props
}) => (
    <nav className={`search-context-bar border-right ${className}`}>
        {results && (
            <>
                <section className="card border-0 rounded-0">
                    <h5 className="card-header rounded-0">Repositories</h5>
                    <ul className="list-group list-group-flush mt-1">
                        {results &&
                            results.dynamicFilters
                                .filter(filter => filter.kind === 'repo' && filter.value !== '')
                                .map((filter, i) => (
                                    <li key={i} className="list-group-item border-0 py-0">
                                        <FilterChip
                                            name={displayRepoName(filter.label)}
                                            query={navbarSearchQuery}
                                            onFilterChosen={onFilterClick}
                                            key={filter.value}
                                            value={filter.value}
                                            count={filter.count}
                                            limitHit={filter.limitHit}
                                        />
                                    </li>
                                ))}
                    </ul>
                </section>
                <section className="card border-0 rounded-0 mt-2">
                    <h5 className="card-header rounded-0">Filters</h5>
                    <ul className="list-group list-group-flush mt-1">
                        {extensionFilters &&
                            extensionFilters
                                .filter(filter => filter.value !== '')
                                .map((filter, i) => (
                                    <li key={i} className="list-group-item border-0 py-0">
                                        <FilterChip
                                            query={navbarSearchQuery}
                                            onFilterChosen={onFilterClick}
                                            value={filter.value}
                                            name={filter.name}
                                        />
                                    </li>
                                ))}
                        {filters
                            .filter(filter => filter.value !== '')
                            .map((filter, i) => (
                                <li key={i} className="list-group-item border-0 py-0">
                                    <FilterChip
                                        query={navbarSearchQuery}
                                        onFilterChosen={onFilterClick}
                                        key={filter.name + filter.value}
                                        value={filter.value}
                                        name={filter.name}
                                    />
                                </li>
                            ))}
                    </ul>
                </section>
                <section className="border-top mt-3 pt-1">
                    <ActionsNavItems
                        {...props}
                        menu={ContributableMenu.SearchResultsToolbar}
                        listClass="flex-column"
                        wrapInList={true}
                        actionItemClass="nav-link px-2"
                    />
                </section>
            </>
        )}
    </nav>
)
