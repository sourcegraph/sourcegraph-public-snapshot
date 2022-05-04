import React, { useMemo } from 'react'

import { useDeepCompareEffectNoCheck } from 'use-deep-compare-effect'
// Disable so we can create a separate store for the VS Code extension.
// eslint-disable-next-line no-restricted-imports
import create from 'zustand'

import {
    InitialParametersSource,
    SearchPatternType,
    SearchQueryState,
    SearchQueryStateStore,
    SearchQueryStateStoreProvider,
    updateQuery,
} from '@sourcegraph/search'
import { SearchSidebar } from '@sourcegraph/search-ui'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { Filter, LATEST_VERSION } from '@sourcegraph/shared/src/search/stream'
import { useObservable } from '@sourcegraph/wildcard'

import { WebviewPageProps } from '../../platform/context'

import styles from './SearchSidebarView.module.scss'

interface SearchSidebarViewProps
    extends Pick<WebviewPageProps, 'settingsCascade' | 'platformContext' | 'extensionCoreAPI'> {
    filters?: Filter[] | undefined
}

export const SearchSidebarView: React.FunctionComponent<React.PropsWithChildren<SearchSidebarViewProps>> = React.memo(
    ({ settingsCascade, platformContext, extensionCoreAPI, filters }) => {
        const useSearchQueryState: SearchQueryStateStore = useMemo(
            () =>
                create<SearchQueryState>((set, get) => ({
                    parametersSource: InitialParametersSource.DEFAULT,
                    queryState: { query: '' },
                    searchCaseSensitivity: false,
                    searchPatternType: SearchPatternType.literal,
                    searchQueryFromURL: '',

                    setQueryState: queryStateUpdate => {
                        const currentSearchQueryState = get()
                        const updatedQueryState =
                            typeof queryStateUpdate === 'function'
                                ? queryStateUpdate(currentSearchQueryState.queryState)
                                : queryStateUpdate

                        extensionCoreAPI
                            .emit({
                                type: 'sidebar_query_update',
                                proposedQueryState: {
                                    queryState: updatedQueryState,
                                    searchCaseSensitivity: currentSearchQueryState.searchCaseSensitivity,
                                    searchPatternType: currentSearchQueryState.searchPatternType,
                                },
                                currentQueryState: {
                                    // Don't spread currentSearchQueryState as it contains un-clone-able functions.
                                    queryState: currentSearchQueryState.queryState,
                                    searchCaseSensitivity: currentSearchQueryState.searchCaseSensitivity,
                                    searchPatternType: currentSearchQueryState.searchPatternType,
                                },
                            })
                            .catch(error => {
                                // TODO surface to user
                                console.error('Error updating search query from Sourcegraph sidebar', error)
                            })

                        extensionCoreAPI.focusSearchPanel().catch(() => {
                            // noop.
                        })
                    },
                    submitSearch: (_submitSearchParameters, updates = []) => {
                        const previousSearchQueryState = get()
                        const updatedQuery = updateQuery(previousSearchQueryState.queryState.query, updates)
                        extensionCoreAPI
                            .streamSearch(updatedQuery, {
                                caseSensitive: previousSearchQueryState.searchCaseSensitivity,
                                patternType: previousSearchQueryState.searchPatternType,
                                version: LATEST_VERSION,
                                trace: undefined,
                            })
                            .catch(error => {
                                // TODO surface to user
                                console.error('Error submitting search from Sourcegraph sidebar', error)
                            })

                        extensionCoreAPI.focusSearchPanel().catch(() => {
                            // noop.
                        })
                    },
                })),
            [extensionCoreAPI]
        )

        const searchQueryStateFromPanel = useObservable(
            useMemo(() => wrapRemoteObservable(extensionCoreAPI.observePanelQueryState()), [extensionCoreAPI])
        )

        useDeepCompareEffectNoCheck(() => {
            if (searchQueryStateFromPanel) {
                useSearchQueryState.setState({
                    queryState: searchQueryStateFromPanel.queryState,
                    searchCaseSensitivity: searchQueryStateFromPanel.searchCaseSensitivity,
                    searchPatternType: searchQueryStateFromPanel.searchPatternType,
                })
            }
        }, [searchQueryStateFromPanel])

        const patternType = useSearchQueryState(state => state.searchPatternType)
        const caseSensitive = useSearchQueryState(state => state.searchCaseSensitivity)

        return (
            <SearchQueryStateStoreProvider useSearchQueryState={useSearchQueryState}>
                <SearchSidebar
                    // Used for SearchTypeLink, which we shouldn't render in the extension.
                    buildSearchURLQueryFromQueryState={() => ''}
                    // Ensure we always render SearchTypeButton which sets zustand state,
                    // instead of URL state which wouldn't make sense in this webview.
                    forceButton={true}
                    caseSensitive={caseSensitive}
                    patternType={patternType}
                    settingsCascade={settingsCascade}
                    telemetryService={platformContext.telemetryService}
                    className={styles.sidebarContainer}
                    filters={filters}
                    // Debt: no selected search context spec
                />
            </SearchQueryStateStoreProvider>
        )
    }
)
