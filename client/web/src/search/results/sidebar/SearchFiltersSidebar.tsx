import { FC, ReactNode, ReactElement, useCallback, useMemo, HTMLAttributes, memo, PropsWithChildren } from 'react'

import { mdiInformationOutline } from '@mdi/js'

import { QueryStateUpdate, QueryUpdate } from '@sourcegraph/search'
import {
    SearchSidebar,
    SearchSidebarSection,
    getDynamicFilterLinks,
    getRepoFilterLinks,
    getSearchSnippetLinks,
    getQuickLinks,
    getSearchReferenceFactory,
    getSearchTypeLinks,
    getFiltersOfKind,
    useLastRepoName,
} from '@sourcegraph/search-ui'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { SectionID } from '@sourcegraph/shared/src/settings/temporary/searchSidebar'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { useCoreWorkflowImprovementsEnabled } from '@sourcegraph/shared/src/settings/useCoreWorkflowImprovementsEnabled'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Code, Tooltip, Icon } from '@sourcegraph/wildcard'

import { buildSearchURLQueryFromQueryState, useExperimentalFeatures } from '../../../stores'
import { AggregationUIMode, GroupResultsPing } from '../components/aggregation'

import { getRevisions } from './Revisions'
import { SearchAggregations } from './SearchAggregations'

export interface SearchFiltersSidebarProps extends TelemetryProps, SettingsCascadeProps, HTMLAttributes<HTMLElement> {
    liveQuery: string
    submittedURLQuery: string
    patternType: SearchPatternType
    caseSensitive: boolean
    filters?: Filter[]
    showAggregationPanel?: boolean
    selectedSearchContextSpec?: string
    aggregationUIMode?: AggregationUIMode
    onNavbarQueryChange: (queryState: QueryStateUpdate) => void
    onSearchSubmit: (updates: QueryUpdate[]) => void
}

export const SearchFiltersSidebar: FC<PropsWithChildren<SearchFiltersSidebarProps>> = memo(props => {
    const {
        liveQuery,
        submittedURLQuery,
        caseSensitive,
        patternType,
        filters,
        showAggregationPanel,
        selectedSearchContextSpec,
        aggregationUIMode,
        onNavbarQueryChange,
        onSearchSubmit,
        telemetryService,
        settingsCascade,
        children,
        ...attributes
    } = props

    // Settings
    const [coreWorkflowImprovementsEnabled] = useCoreWorkflowImprovementsEnabled()
    const enableSearchAggregations = useExperimentalFeatures(features => features.searchResultsAggregations ?? true)
    const proactiveSearchAggregations = useExperimentalFeatures(
        features => features.proactiveSearchResultsAggregations ?? true
    )
    const [, setSelectedTab] = useTemporarySetting('search.sidebar.selectedTab', 'filters')

    // Derived state
    const repoFilters = useMemo(() => getFiltersOfKind(filters, FilterType.repo), [filters])
    const repoName = useLastRepoName(liveQuery, repoFilters)

    const onDynamicFilterClicked = useCallback(
        (value: string, kind?: string) => {
            telemetryService.log('DynamicFilterClicked', { search_filter: { kind } })
            onSearchSubmit([{ type: 'toggleSubquery', value }])
        },
        [telemetryService, onSearchSubmit]
    )

    const onSnippetClicked = useCallback(
        (value: string) => {
            telemetryService.log('SearchSnippetClicked')
            onSearchSubmit([{ type: 'toggleSubquery', value }])
        },
        [telemetryService, onSearchSubmit]
    )

    const handleAggregationBarLinkClick = useCallback(
        (query: string): void => {
            onSearchSubmit([{ type: 'replaceQuery', value: query }])
        },
        [onSearchSubmit]
    )

    const handleGroupedByToggle = useCallback(
        (open: boolean): void => {
            telemetryService.log(open ? GroupResultsPing.ExpandSidebarSection : GroupResultsPing.CollapseSidebarSection)
        },
        [telemetryService]
    )

    return (
        <SearchSidebar {...attributes} onClose={() => setSelectedTab(null)}>
            {children}

            {showAggregationPanel && enableSearchAggregations && aggregationUIMode === AggregationUIMode.Sidebar && (
                <SearchSidebarSection
                    sectionId={SectionID.GROUPED_BY}
                    header={<CustomAggregationHeading telemetryService={props.telemetryService} />}
                    // SearchAggregations content contains component that makes a few API network requests
                    // in order to prevent these calls if this section is collapsed we turn off force render
                    // for collapse section component
                    forcedRender={false}
                    onToggle={handleGroupedByToggle}
                >
                    <SearchAggregations
                        query={submittedURLQuery}
                        patternType={patternType}
                        proactive={proactiveSearchAggregations}
                        caseSensitive={caseSensitive}
                        telemetryService={telemetryService}
                        onQuerySubmit={handleAggregationBarLinkClick}
                    />
                </SearchSidebarSection>
            )}

            <SearchSidebarSection sectionId={SectionID.SEARCH_TYPES} header="Search Types">
                {getSearchTypeLinks({
                    query: liveQuery,
                    onNavbarQueryChange,
                    selectedSearchContextSpec,
                    buildSearchURLQueryFromQueryState,
                    forceButton: false,
                })}
            </SearchSidebarSection>

            {!coreWorkflowImprovementsEnabled && (
                <SearchSidebarSection sectionId={SectionID.DYNAMIC_FILTERS} header="Dynamic Filters">
                    {getDynamicFilterLinks(
                        filters,
                        ['lang', 'file', 'utility'],
                        onDynamicFilterClicked,
                        (label, value) => `Filter by ${value}`,
                        (label, value) => value
                    )}
                </SearchSidebarSection>
            )}

            {coreWorkflowImprovementsEnabled && (
                <SearchSidebarSection sectionId={SectionID.LANGUAGES} header="Languages">
                    {getDynamicFilterLinks(filters, ['lang'], onDynamicFilterClicked, label => `Search ${label} files`)}
                </SearchSidebarSection>
            )}

            <SearchSidebarSection
                sectionId={SectionID.REPOSITORIES}
                header="Repositories"
                showSearch={true}
                minItems={1}
                noResultText={getRepoFilterNoResultText}
            >
                {getRepoFilterLinks(repoFilters, onDynamicFilterClicked, coreWorkflowImprovementsEnabled)}
            </SearchSidebarSection>

            {coreWorkflowImprovementsEnabled && (
                <>
                    <SearchSidebarSection sectionId={SectionID.FILE_TYPES} header="File types">
                        {getDynamicFilterLinks(filters, ['file'], onDynamicFilterClicked)}
                    </SearchSidebarSection>
                    <SearchSidebarSection sectionId={SectionID.OTHER} header="Other">
                        {getDynamicFilterLinks(filters, ['utility'], onDynamicFilterClicked)}
                    </SearchSidebarSection>
                </>
            )}

            {repoName && (
                <SearchSidebarSection
                    sectionId={SectionID.REVISIONS}
                    header="Revisions"
                    showSearch={true}
                    clearSearchOnChange={repoName}
                >
                    {getRevisions({ repoName, onFilterClick: onSearchSubmit })}
                </SearchSidebarSection>
            )}

            <SearchSidebarSection
                sectionId={SectionID.SEARCH_REFERENCE}
                header="Search reference"
                showSearch={true}
                // search reference should always preserve the filter
                // (false is just an arbitrary but static value)
                clearSearchOnChange={false}
            >
                {getSearchReferenceFactory({ telemetryService, setQueryState: onNavbarQueryChange })}
            </SearchSidebarSection>

            <SearchSidebarSection sectionId={SectionID.SEARCH_SNIPPETS} header="Search snippets">
                {getSearchSnippetLinks(settingsCascade, onSnippetClicked)}
            </SearchSidebarSection>

            {!coreWorkflowImprovementsEnabled && (
                <SearchSidebarSection sectionId={SectionID.QUICK_LINKS} header="Quicklinks">
                    {getQuickLinks(settingsCascade)}
                </SearchSidebarSection>
            )}
        </SearchSidebar>
    )
})

SearchFiltersSidebar.displayName = 'SearchFiltersSidebar'

const getRepoFilterNoResultText = (repoFilterLinks: ReactElement[]): ReactNode => (
    <span>
        None of the top {repoFilterLinks.length} repositories in your results match this filter. Try a{' '}
        <Code>repo:</Code> search in the main search bar instead.
    </span>
)

const CustomAggregationHeading: FC<TelemetryProps> = ({ telemetryService }) => (
    <>
        Group results by
        <Tooltip content="Aggregation is based on results with no count limitation (count:all).">
            <Icon
                aria-label="Info icon about aggregation run"
                size="md"
                svgPath={mdiInformationOutline}
                onMouseEnter={() => telemetryService.log(GroupResultsPing.InfoIconHover)}
            />
        </Tooltip>
    </>
)
