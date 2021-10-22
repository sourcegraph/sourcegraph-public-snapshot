import classNames from 'classnames'
import React, { useCallback, useMemo } from 'react'
import { useHistory } from 'react-router'
import StickyBox from 'react-sticky-box'
import shallow from 'zustand/shallow'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { TemporarySettings } from '../../../settings/temporary/TemporarySettings'
import { useTemporarySetting } from '../../../settings/temporary/useTemporarySetting'
import { SubmitSearchParameters } from '../../helpers'
import { NavbarQueryState, QueryUpdate, useNavbarQueryState } from '../../navbarSearchQueryState'

import { getDynamicFilterLinks, getRepoFilterLinks, getSearchSnippetLinks } from './FilterLink'
import { getFiltersOfKind, useLastRepoName } from './helpers'
import { getQuickLinks } from './QuickLink'
import { getRevisions } from './Revisions'
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
}: NavbarQueryState): {
    query: string
    setQueryState: NavbarQueryState['setQueryState']
    submitSearch: NavbarQueryState['submitSearch']
} => ({
    query,
    setQueryState,
    submitSearch,
})

export const SearchSidebar: React.FunctionComponent<SearchSidebarProps> = props => {
    const history = useHistory()
    const [collapsedSections, setCollapsedSections] = useTemporarySetting('search.collapsedSidebarSections', {})
    const { query, setQueryState, submitSearch } = useNavbarQueryState(selectFromQueryState, shallow)

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
                {repoName ? (
                    <SearchSidebarSection
                        sectionId={SectionID.REVISIONS}
                        className={styles.searchSidebarItem}
                        header="Revisions"
                        startCollapsed={collapsedSections?.[SectionID.REVISIONS]}
                        onToggle={persistToggleState}
                        showSearch={true}
                        clearSearchOnChange={repoName}
                    >
                        {getRevisions({ repoName, onFilterClick: submitQueryWithProps })}
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
