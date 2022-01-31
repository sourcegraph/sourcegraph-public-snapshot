import React, { useMemo } from 'react'
import create from 'zustand'

import {
    SearchPatternType,
    SearchQueryState,
    SearchQueryStateStore,
    SearchQueryStateStoreProvider,
} from '@sourcegraph/search'
import { SearchSidebar } from '@sourcegraph/search-ui/src/results/sidebar/SearchSidebar'

import { WebviewPageProps } from '../platform/context'

import styles from './SearchSidebarView.module.scss'

interface SearchSidebarViewProps extends WebviewPageProps {}

export const SearchSidebarView: React.FunctionComponent<SearchSidebarViewProps> = ({
    settingsCascade,
    platformContext,
}) => {
    const useSearchQueryState: SearchQueryStateStore = useMemo(
        () =>
            create<SearchQueryState>((set, get) => ({
                queryState: { query: '' },
                searchCaseSensitivity: false,
                searchPatternType: SearchPatternType.literal,
                searchQueryFromURL: '',

                setQueryState: () => {},
                submitSearch: () => {},
            })),
        []
    )

    return (
        <SearchQueryStateStoreProvider useSearchQueryState={useSearchQueryState}>
            <SearchSidebar
                buildSearchURLQueryFromQueryState={() => ''}
                caseSensitive={true}
                patternType={SearchPatternType.literal}
                settingsCascade={settingsCascade}
                telemetryService={platformContext.telemetryService}
                className={styles.sidebarContainer}
            />
        </SearchQueryStateStoreProvider>
    )
}
