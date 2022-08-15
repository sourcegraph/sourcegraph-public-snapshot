import React, { FC, ReactNode, ReactElement, useCallback, useMemo, HTMLAttributes, memo } from 'react'

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
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { SectionID } from '@sourcegraph/shared/src/settings/temporary/searchSidebar'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { useCoreWorkflowImprovementsEnabled } from '@sourcegraph/shared/src/settings/useCoreWorkflowImprovementsEnabled'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Code } from '@sourcegraph/wildcard'

import { buildSearchURLQueryFromQueryState } from '../../../stores'

import { getRevisions } from './Revisions'

export interface SearchFiltersSidebarProps extends TelemetryProps, SettingsCascadeProps, HTMLAttributes<HTMLElement> {
    query: string
    onNavbarQueryChange: (queryState: QueryStateUpdate) => void
    onSearchSubmit: (updates: QueryUpdate[]) => void

    filters?: Filter[]
    selectedSearchContextSpec?: string
    /** Content to render inside sidebar, but before other sections. */
    prefixContent?: React.ReactNode
}

export const SearchFiltersSidebar: FC<SearchFiltersSidebarProps> = memo(props => {
    const {
        query,
        filters,
        selectedSearchContextSpec,
        telemetryService,
        settingsCascade,
        prefixContent,
        onNavbarQueryChange,
        onSearchSubmit,
        ...attributes
    } = props

    const [coreWorkflowImprovementsEnabled] = useCoreWorkflowImprovementsEnabled()
    const [, setSelectedTab] = useTemporarySetting('search.sidebar.selectedTab', 'filters')

    const repoFilters = useMemo(() => getFiltersOfKind(filters, FilterType.repo), [filters])
    const repoName = useLastRepoName(query, repoFilters)

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

    return (
        <SearchSidebar onClose={() => setSelectedTab(null)} {...attributes}>
            {prefixContent}

            <SearchSidebarSection sectionId={SectionID.SEARCH_TYPES} header="Search Types">
                {getSearchTypeLinks({
                    query,
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

            {!coreWorkflowImprovementsEnabled && (
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
            )}

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

const getRepoFilterNoResultText = (repoFilterLinks: ReactElement[]): ReactNode => (
    <span>
        None of the top {repoFilterLinks.length} repositories in your results match this filter. Try a{' '}
        <Code>repo:</Code> search in the main search bar instead.
    </span>
)
