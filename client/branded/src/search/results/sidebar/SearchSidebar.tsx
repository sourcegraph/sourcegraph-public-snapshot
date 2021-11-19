import classNames from 'classnames'
import React, { useCallback, useMemo } from 'react'
import { useHistory } from 'react-router'
import StickyBox from 'react-sticky-box'
import { UseStore } from 'zustand'
import shallow from 'zustand/shallow'

import { SubmitSearchParameters } from '@sourcegraph/shared/src/search/helpers'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { QueryUpdate, SearchQueryState } from '@sourcegraph/shared/src/search/searchQueryState'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TemporarySettings } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { getDynamicFilterLinks, getRepoFilterLinks, getSearchSnippetLinks } from './FilterLink'
import { getFiltersOfKind, useLastRepoName } from './helpers'
import { getQuickLinks } from './QuickLink'
import { RevisionsProps } from './Revisions'
import { getSearchReferenceFactory } from './SearchReference'
import styles from './SearchSidebar.module.scss'
import { SearchSidebarSection } from './SearchSidebarSection'
import { getSearchTypeLinks } from './SearchTypeLink'

export interface SearchSidebarProps
    extends Omit<SubmitSearchParameters, 'history' | 'query' | 'source' | 'searchParameters'>,
        SettingsCascadeProps,
        TelemetryProps {
    filters?: Filter[]
    className?: string

    /**
     * Not yet implemented in the VS Code extension (blocked on Apollo Client integration).
     * */
    getRevisions?: (revisionsProps: Omit<RevisionsProps, 'query'>) => (query: string) => JSX.Element

    /**
     * Zustand store. Passed as a prop because there may be different global stores across clients
     * (e.g. VS Code extension, web app), so the sidebar expects a store with the minimal interface
     * for search.
     */
    useQueryState: UseStore<SearchQueryState>

    /**
     * Force search type links to be rendered as buttons.
     * Used e.g. in the VS Code extension to update search query state.
     * */
    forceButton?: boolean
}

export enum SectionID {
    SEARCH_REFERENCE = 'reference',
    SEARCH_TYPES = 'types',
    DYNAMIC_FILTERS = 'filters',
    REPOSITORIES = 'repositories',
    SEARCH_SNIPPETS = 'snippets',
    QUICK_LINKS = 'quicklinks',
    REVISIONS = 'revisions',
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

export const SearchSidebar: React.FunctionComponent<SearchSidebarProps> = props => {
    const history = useHistory()
    const [collapsedSections, setCollapsedSections] = useTemporarySetting('search.collapsedSidebarSections', {})
    const { query, setQueryState, submitSearch } = props.useQueryState(selectFromQueryState, shallow)

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
                    [id]: open,
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
                        caseSensitive: props.caseSensitive,
                        onNavbarQueryChange: setQueryState,
                        patternType: props.patternType,
                        query,
                        selectedSearchContextSpec: props.selectedSearchContextSpec,
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
                {repoName && props.getRevisions ? (
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

    return <div className={classNames(styles.searchSidebar, props.className)}>{body}</div>
}
