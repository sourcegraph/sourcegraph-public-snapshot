import React, { useCallback, useMemo } from 'react'

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
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { getDynamicFilterLinks, getRepoFilterLinks, getSearchSnippetLinks } from './FilterLink'
import { getFiltersOfKind, useLastRepoName } from './helpers'
import { getQuickLinks } from './QuickLink'
import { RevisionsProps } from './revisions'
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
    showOnboardingTour?: boolean

    /**
     * Not yet implemented in the VS Code extension (blocked on Apollo Client integration).
     */
    getRevisions?: (revisionsProps: Omit<RevisionsProps, 'query'>) => (query: string) => JSX.Element

    /**
     * Content to render inside sidebar, but before other sections.
     */
    prefixContent?: JSX.Element

    buildSearchURLQueryFromQueryState: (queryParameters: BuildSearchQueryURLParameters) => string

    /**
     * Force search type links to be rendered as buttons.
     * Used e.g. in the VS Code extension to update search query state.
     */
    forceButton?: boolean
}

const selectFromQueryState = ({
    queryState: { query },
    setQueryState,
    submitSearch,
}: SearchQueryState): {
    query: string
    setQueryState: SearchQueryState['setQueryState']
    submitSearch: SearchQueryState['submitSearch']
} => ({
    query,
    setQueryState,
    submitSearch,
})

export const SearchSidebar: React.FunctionComponent<React.PropsWithChildren<SearchSidebarProps>> = props => {
    const history = useHistory()
    const [collapsedSections, setCollapsedSections] = useTemporarySetting('search.collapsedSidebarSections', {})

    // The zustand store for search query state is referenced through context
    // because there may be different global stores across clients
    // (e.g. VS Code extension, web app)
    const { query, setQueryState, submitSearch } = useSearchQueryStateStoreContext()(selectFromQueryState, shallow)

    // Unlike onFilterClicked, this function will always append or update a filter
    const submitQueryWithProps = useCallback(
        (updates: QueryUpdate[]) => submitSearch({ ...props, source: 'filter', history }, updates),
        [history, props, submitSearch]
    )

    const onDynamicFilterClicked = useCallback(
        (value: string) => {
            props.telemetryService.log('DynamicFilterClicked', {
                search_filter: { value },
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
    const repoFilterLinks = useMemo(() => getRepoFilterLinks(repoFilters, onDynamicFilterClicked), [
        repoFilters,
        onDynamicFilterClicked,
    ])
    const dynamicFilterLinks = useMemo(() => getDynamicFilterLinks(props.filters, onDynamicFilterClicked), [
        props.filters,
        onDynamicFilterClicked,
    ])
    const showReposSection = repoFilterLinks.length > 1

    let body

    // collapsedSections is undefined on first render. To prevent the sections
    // being rendered open and immediately closing them, we render them only after
    // we got the settings.
    if (collapsedSections) {
        body = (
            <StickyBox className={styles.searchSidebarStickyBox}>
                <SearchSidebarSection
                    sectionId={SectionID.SEARCH_TYPES}
                    className={styles.searchSidebarItem}
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
                <SearchSidebarSection
                    sectionId={SectionID.DYNAMIC_FILTERS}
                    className={styles.searchSidebarItem}
                    header="Dynamic filters"
                    startCollapsed={collapsedSections?.[SectionID.DYNAMIC_FILTERS]}
                    onToggle={persistToggleState}
                >
                    {dynamicFilterLinks}
                </SearchSidebarSection>
                {showReposSection ? (
                    <SearchSidebarSection
                        sectionId={SectionID.REPOSITORIES}
                        className={styles.searchSidebarItem}
                        header="Repositories"
                        startCollapsed={collapsedSections?.[SectionID.REPOSITORIES]}
                        onToggle={persistToggleState}
                        showSearch={true}
                        noResultText={
                            <span>
                                None of the top {repoFilterLinks.length} repositories in your results match this filter.
                                Try a <code>repo:</code> search in the main search bar instead.
                            </span>
                        }
                    >
                        {repoFilterLinks}
                    </SearchSidebarSection>
                ) : null}
                {props.getRevisions && repoName ? (
                    <SearchSidebarSection
                        sectionId={SectionID.REVISIONS}
                        className={styles.searchSidebarItem}
                        header="Revisions"
                        startCollapsed={collapsedSections?.[SectionID.REVISIONS]}
                        onToggle={persistToggleState}
                        showSearch={true}
                        clearSearchOnChange={repoName}
                    >
                        {props.getRevisions({ repoName, onFilterClick: submitQueryWithProps })}
                    </SearchSidebarSection>
                ) : null}
                <SearchSidebarSection
                    sectionId={SectionID.SEARCH_REFERENCE}
                    className={styles.searchSidebarItem}
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
                <SearchSidebarSection
                    sectionId={SectionID.SEARCH_SNIPPETS}
                    className={styles.searchSidebarItem}
                    header="Search snippets"
                    startCollapsed={collapsedSections?.[SectionID.SEARCH_SNIPPETS]}
                    onToggle={persistToggleState}
                >
                    {getSearchSnippetLinks(props.settingsCascade, onSnippetClicked)}
                </SearchSidebarSection>
                <SearchSidebarSection
                    sectionId={SectionID.QUICK_LINKS}
                    className={styles.searchSidebarItem}
                    header="Quicklinks"
                    startCollapsed={collapsedSections?.[SectionID.QUICK_LINKS]}
                    onToggle={persistToggleState}
                >
                    {getQuickLinks(props.settingsCascade)}
                </SearchSidebarSection>
            </StickyBox>
        )
    }

    return (
        <div className={classNames(styles.searchSidebar, props.className)}>
            {props.prefixContent}
            {body}
        </div>
    )
}
