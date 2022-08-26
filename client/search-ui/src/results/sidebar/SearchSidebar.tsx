import React, { useCallback, useMemo } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'
import { useHistory } from 'react-router'
import StickyBox from 'react-sticky-box'
import shallow from 'zustand/shallow'

import {
    BuildSearchQueryURLParameters,
    QueryUpdate,
    SearchQueryState,
    SubmitSearchParameters,
    useSearchQueryStateStoreContext,
} from '@sourcegraph/search'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { SectionID } from '@sourcegraph/shared/src/settings/temporary/searchSidebar'
import { TemporarySettings } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { useCoreWorkflowImprovementsEnabled } from '@sourcegraph/shared/src/settings/useCoreWorkflowImprovementsEnabled'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Code, Icon } from '@sourcegraph/wildcard'

import { AggregationUIMode, useAggregationUIMode } from '../aggregation'

import { getDynamicFilterLinks, getRepoFilterLinks, getSearchSnippetLinks } from './FilterLink'
import { getFiltersOfKind, useLastRepoName } from './helpers'
import { getQuickLinks } from './QuickLink'
import { RevisionsProps } from './revisions'
import { SearchAggregations } from './SearchAggregations'
import { getSearchReferenceFactory } from './SearchReference'
import { SearchSidebarSection } from './SearchSidebarSection'
import { getSearchTypeLinks } from './SearchTypeLink'

import styles from './SearchSidebar.module.scss'

export interface SearchSidebarProps
    extends Omit<SubmitSearchParameters, 'history' | 'query' | 'source' | 'searchParameters'>,
        SettingsCascadeProps,
        TelemetryProps {
    filters?: Filter[]
    className?: string

    /**
     * Not yet implemented in the VS Code extension (blocked on Apollo Client integration).
     */
    getRevisions?: (revisionsProps: Omit<RevisionsProps, 'query'>) => (query: string) => React.ReactNode

    /**
     * Content to render inside sidebar, but before other sections.
     */
    prefixContent?: React.ReactNode

    buildSearchURLQueryFromQueryState: (queryParameters: BuildSearchQueryURLParameters) => string

    /**
     * Force search type links to be rendered as buttons.
     * Used e.g. in the VS Code extension to update search query state.
     */
    forceButton?: boolean

    /**
     * Enables search compute-based aggregations filter panel
     */
    enableSearchAggregation?: boolean
}

const selectFromQueryState = ({
    queryState: { query },
    setQueryState,
    submitSearch,
    searchQueryFromURL,
    searchPatternType,
}: SearchQueryState): {
    query: string
    setQueryState: SearchQueryState['setQueryState']
    submitSearch: SearchQueryState['submitSearch']
    searchQueryFromURL: SearchQueryState['searchQueryFromURL']
    searchPatternType: SearchQueryState['searchPatternType']
} => ({
    query,
    setQueryState,
    submitSearch,
    searchQueryFromURL,
    searchPatternType,
})

export const SearchSidebar: React.FunctionComponent<SearchSidebarProps> = props => {
    const history = useHistory()
    const [collapsedSections, setCollapsedSections] = useTemporarySetting('search.collapsedSidebarSections', {})
    const [, setSelectedTab] = useTemporarySetting('search.sidebar.selectedTab', 'filters')
    const [coreWorkflowImprovementsEnabled] = useCoreWorkflowImprovementsEnabled()

    // The zustand store for search query state is referenced through context
    // because there may be different global stores across clients
    // (e.g. VS Code extension, web app)
    const {
        query,
        searchQueryFromURL,
        searchPatternType,
        setQueryState,
        submitSearch,
    } = useSearchQueryStateStoreContext()(selectFromQueryState, shallow)
    const [aggregationUIMode] = useAggregationUIMode()

    // Unlike onFilterClicked, this function will always append or update a filter
    const submitQueryWithProps = useCallback(
        (updates: QueryUpdate[]) => submitSearch({ ...props, source: 'filter', history }, updates),
        [history, props, submitSearch]
    )

    const onDynamicFilterClicked = useCallback(
        (value: string, kind?: string) => {
            props.telemetryService.log('DynamicFilterClicked', {
                search_filter: { kind },
            })

            submitQueryWithProps([{ type: 'toggleSubquery', value }])
        },
        [submitQueryWithProps, props.telemetryService]
    )

    const onSnippetClicked = useCallback(
        (value: string) => {
            props.telemetryService.log('SearchSnippetClicked')
            submitQueryWithProps([{ type: 'toggleSubquery', value }])
        },
        [submitQueryWithProps, props.telemetryService]
    )

    const persistToggleState = useCallback(
        (id: string, open: boolean) => {
            setCollapsedSections(openSections => {
                const newSettings: TemporarySettings['search.collapsedSidebarSections'] = {
                    ...openSections,
                    [id]: !open,
                }
                return newSettings
            })
        },
        [setCollapsedSections]
    )
    const onSearchReferenceToggle = useCallback(
        (_id: string, open: boolean) => {
            persistToggleState(SectionID.SEARCH_REFERENCE, open)
            props.telemetryService.log(open ? 'SearchReferenceOpened' : 'SearchReferenceClosed')
        },
        [persistToggleState, props.telemetryService]
    )

    const repoFilters = useMemo(() => getFiltersOfKind(props.filters, FilterType.repo), [props.filters])
    const repoName = useLastRepoName(query, repoFilters)
    const repoFilterLinks = useMemo(
        () => getRepoFilterLinks(repoFilters, onDynamicFilterClicked, coreWorkflowImprovementsEnabled),
        [repoFilters, onDynamicFilterClicked, coreWorkflowImprovementsEnabled]
    )
    const showReposSection = repoFilterLinks.length > 1

    const langFilterLinks = useMemo(
        () => getDynamicFilterLinks(props.filters, ['lang'], onDynamicFilterClicked, label => `Search ${label} files`),
        [props.filters, onDynamicFilterClicked]
    )
    const fileFilterLinks = useMemo(() => getDynamicFilterLinks(props.filters, ['file'], onDynamicFilterClicked), [
        props.filters,
        onDynamicFilterClicked,
    ])
    const utilityFilterLinks = useMemo(
        () => getDynamicFilterLinks(props.filters, ['utility'], onDynamicFilterClicked),
        [props.filters, onDynamicFilterClicked]
    )
    const dynamicFilterLinks = useMemo(
        () =>
            getDynamicFilterLinks(
                props.filters,
                ['lang', 'file', 'utility'],
                onDynamicFilterClicked,
                (label, value) => `Filter by ${value}`,
                (label, value) => value
            ),
        [props.filters, onDynamicFilterClicked]
    )

    const handleAggregationBarLinkClick = (query: string): void => {
        submitQueryWithProps([{ type: 'replaceQuery', value: query }])
    }

    let body

    // collapsedSections is undefined on first render. To prevent the sections
    // being rendered open and immediately closing them, we render them only after
    // we got the settings.
    if (collapsedSections) {
        body = (
            <>
                {props.enableSearchAggregation && aggregationUIMode === AggregationUIMode.Sidebar && (
                    <SearchSidebarSection
                        sectionId={SectionID.GROUPED_BY}
                        className={styles.item}
                        header="Group results by"
                        startCollapsed={collapsedSections?.[SectionID.GROUPED_BY]}
                        onToggle={persistToggleState}
                    >
                        <SearchAggregations
                            query={searchQueryFromURL}
                            patternType={searchPatternType}
                            onQuerySubmit={handleAggregationBarLinkClick}
                        />
                    </SearchSidebarSection>
                )}

                <SearchSidebarSection
                    sectionId={SectionID.SEARCH_TYPES}
                    className={styles.item}
                    header="Search Types"
                    startCollapsed={collapsedSections?.[SectionID.SEARCH_TYPES]}
                    onToggle={persistToggleState}
                >
                    {getSearchTypeLinks({
                        onNavbarQueryChange: setQueryState,
                        query,
                        selectedSearchContextSpec: props.selectedSearchContextSpec,
                        buildSearchURLQueryFromQueryState: props.buildSearchURLQueryFromQueryState,
                        forceButton: props.forceButton,
                    })}
                </SearchSidebarSection>
                {!coreWorkflowImprovementsEnabled && dynamicFilterLinks.length > 0 && (
                    <SearchSidebarSection
                        sectionId={SectionID.DYNAMIC_FILTERS}
                        className={styles.item}
                        header="Dynamic Filters"
                        startCollapsed={collapsedSections?.[SectionID.DYNAMIC_FILTERS]}
                        onToggle={persistToggleState}
                    >
                        {dynamicFilterLinks}
                    </SearchSidebarSection>
                )}
                {coreWorkflowImprovementsEnabled && langFilterLinks.length > 0 && (
                    <SearchSidebarSection
                        sectionId={SectionID.LANGUAGES}
                        className={styles.item}
                        header="Languages"
                        startCollapsed={collapsedSections?.[SectionID.LANGUAGES]}
                        onToggle={persistToggleState}
                    >
                        {langFilterLinks}
                    </SearchSidebarSection>
                )}
                {showReposSection && (
                    <SearchSidebarSection
                        sectionId={SectionID.REPOSITORIES}
                        className={styles.item}
                        header="Repositories"
                        startCollapsed={collapsedSections?.[SectionID.REPOSITORIES]}
                        onToggle={persistToggleState}
                        showSearch={true}
                        noResultText={
                            <span>
                                None of the top {repoFilterLinks.length} repositories in your results match this filter.
                                Try a <Code>repo:</Code> search in the main search bar instead.
                            </span>
                        }
                    >
                        {repoFilterLinks}
                    </SearchSidebarSection>
                )}
                {coreWorkflowImprovementsEnabled && fileFilterLinks.length > 0 && (
                    <SearchSidebarSection
                        sectionId={SectionID.FILE_TYPES}
                        className={styles.item}
                        header="File types"
                        startCollapsed={collapsedSections?.[SectionID.FILE_TYPES]}
                        onToggle={persistToggleState}
                    >
                        {fileFilterLinks}
                    </SearchSidebarSection>
                )}
                {coreWorkflowImprovementsEnabled && utilityFilterLinks.length > 0 && (
                    <SearchSidebarSection
                        sectionId={SectionID.OTHER}
                        className={styles.item}
                        header="Other"
                        startCollapsed={collapsedSections?.[SectionID.OTHER]}
                        onToggle={persistToggleState}
                    >
                        {utilityFilterLinks}
                    </SearchSidebarSection>
                )}
                {props.getRevisions && repoName ? (
                    <SearchSidebarSection
                        sectionId={SectionID.REVISIONS}
                        className={styles.item}
                        header="Revisions"
                        startCollapsed={collapsedSections?.[SectionID.REVISIONS]}
                        onToggle={persistToggleState}
                        showSearch={true}
                        clearSearchOnChange={repoName}
                    >
                        {props.getRevisions({ repoName, onFilterClick: submitQueryWithProps })}
                    </SearchSidebarSection>
                ) : null}
                {!coreWorkflowImprovementsEnabled && (
                    <SearchSidebarSection
                        sectionId={SectionID.SEARCH_REFERENCE}
                        className={styles.item}
                        header="Search reference"
                        showSearch={true}
                        startCollapsed={collapsedSections?.[SectionID.SEARCH_REFERENCE]}
                        onToggle={onSearchReferenceToggle}
                        // search reference should always preserve the filter
                        // (false is just an arbitrary but static value)
                        clearSearchOnChange={false}
                    >
                        {getSearchReferenceFactory({
                            telemetryService: props.telemetryService,
                            setQueryState,
                        })}
                    </SearchSidebarSection>
                )}
                <SearchSidebarSection
                    sectionId={SectionID.SEARCH_SNIPPETS}
                    className={styles.item}
                    header="Search snippets"
                    startCollapsed={collapsedSections?.[SectionID.SEARCH_SNIPPETS]}
                    onToggle={persistToggleState}
                >
                    {getSearchSnippetLinks(props.settingsCascade, onSnippetClicked)}
                </SearchSidebarSection>
                {!coreWorkflowImprovementsEnabled && (
                    <SearchSidebarSection
                        sectionId={SectionID.QUICK_LINKS}
                        className={styles.item}
                        header="Quicklinks"
                        startCollapsed={collapsedSections?.[SectionID.QUICK_LINKS]}
                        onToggle={persistToggleState}
                    >
                        {getQuickLinks(props.settingsCascade)}
                    </SearchSidebarSection>
                )}
            </>
        )
    }

    return (
        <aside className={classNames(styles.sidebar, props.className)} role="region" aria-label="Search sidebar">
            <StickyBox className={styles.stickyBox} offsetTop={8}>
                <div className={styles.header}>
                    <Button variant="icon" onClick={() => setSelectedTab(null)}>
                        <Icon svgPath={mdiClose} aria-label="Close sidebar" />
                    </Button>
                </div>
                {props.prefixContent}
                {body}
            </StickyBox>
        </aside>
    )
}
