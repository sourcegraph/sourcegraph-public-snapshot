import React, { useEffect, useMemo, useState } from 'react'
import create, { UseStore } from 'zustand'

import { SearchSidebar as BrandedSearchSidebar } from '@sourcegraph/branded/src/search/results/sidebar/SearchSidebar'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { SearchQueryState, updateQuery } from '@sourcegraph/shared/src/search/searchQueryState'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { CurrentAuthStateResult, CurrentAuthStateVariables, SearchPatternType } from '../../graphql-operations'
import { LocalRecentSeachProps } from '../contract'
import { WebviewPageProps } from '../platform/context'
import { currentAuthStateQuery } from '../search-panel/queries'

import { HistorySidebar } from './HistorySidebar'
import { OpenSearchPanelCta } from './OpenSearchPanelCta'
import styles from './SearchSidebar.module.scss'
import { SidebarAuthCheck } from './SidebarAuthCheck'
interface SearchSidebarProps extends WebviewPageProps {}

export const SearchSidebar: React.FC<SearchSidebarProps> = ({
    sourcegraphVSCodeExtensionAPI,
    platformContext,
    theme,
}) => {
    // Check if there is any opened / active search panel
    const [activeSearchPanel, setActiveSearchPanel] = useState<boolean | undefined>(undefined)
    const [localRecentSearches, setLocalRecentSearches] = useState<LocalRecentSeachProps[] | undefined>(undefined)
    const [validating, setValidating] = useState(true)
    // Check if User is currently on VS Code Desktop or VS Code Web
    const [onDesktop, setOnDesktop] = useState<boolean | undefined>(undefined)
    const [corsUri, setCorsUri] = useState<string | undefined>(undefined)
    const [hasAccount, setHasAccount] = useState(false)
    const [hasAccessToken, setHasAccessToken] = useState<boolean | undefined>(undefined)
    const [validAccessToken, setValidAccessToken] = useState<boolean | undefined>(undefined)
    const [authenticatedUser, setAuthenticatedUser] = useState<AuthenticatedUser | null>(null)
    // Search Query
    const [patternType, setPatternType] = useState<SearchPatternType>(SearchPatternType.literal)
    const [caseSensitive, setCaseSensitive] = useState<boolean>(false)

    // Get current access token, cros, and platform settings
    useEffect(() => {
        setValidating(true)
        // Get initial settings
        if (onDesktop === undefined) {
            // Get Local Search History
            sourcegraphVSCodeExtensionAPI
                .getLocalRecentSearch()
                .then(response => {
                    setLocalRecentSearches(response)
                })
                .catch(() => {
                    // TODO error handling
                })
            // Get Current platform
            sourcegraphVSCodeExtensionAPI
                .onDesktop()
                .then(isOnDesktop => setOnDesktop(isOnDesktop))
                .catch(error => console.error(error))
            // Get Cors from Setting
            sourcegraphVSCodeExtensionAPI
                .getCorsSetting()
                .then(uri => {
                    setCorsUri(uri)
                })
                .catch(error => console.error(error))
            // Get Access Token from Setting
            sourcegraphVSCodeExtensionAPI
                .hasAccessToken()
                .then(response => {
                    setHasAccessToken(response)
                    setHasAccount(response)
                })
                .catch(() => setHasAccessToken(false))
        }
        if (hasAccessToken === false) {
            setValidAccessToken(false)
        }
        // Validate Token
        if (hasAccount && hasAccessToken && validAccessToken === undefined) {
            ;(async () => {
                const currentUser = await platformContext
                    .requestGraphQL<CurrentAuthStateResult, CurrentAuthStateVariables>({
                        request: currentAuthStateQuery,
                        variables: {},
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()
                // If user is detected, set valid access token to true
                if (currentUser.data) {
                    setAuthenticatedUser(currentUser.data.currentUser)
                } else {
                    setAuthenticatedUser(null)
                }
            })().catch(() => setValidAccessToken(false))
        }
        setValidating(false)
    }, [
        sourcegraphVSCodeExtensionAPI,
        onDesktop,
        corsUri,
        hasAccessToken,
        validAccessToken,
        hasAccount,
        platformContext,
        setHasAccessToken,
    ])

    const useQueryState: UseStore<SearchQueryState> = useMemo(() => {
        const useStore = create<SearchQueryState>((set, get) => ({
            queryState: { query: '' },
            setQueryState: async queryStateUpdate => {
                const queryState =
                    typeof queryStateUpdate === 'function' ? queryStateUpdate(get().queryState) : queryStateUpdate
                set({ queryState })
                // TODO error handling

                await sourcegraphVSCodeExtensionAPI.setActiveWebviewQueryState(queryState)
            },
            submitSearch: (_parameters, updates = []) => {
                const updatedQuery = updateQuery(get().queryState.query, updates)
                // TODO error handling
                sourcegraphVSCodeExtensionAPI
                    .submitActiveWebviewSearch({
                        query: updatedQuery,
                    })
                    .then(
                        () => {
                            setActiveSearchPanel(true)
                        },
                        () => {}
                    )
            },
        }))
        return useStore
    }, [sourcegraphVSCodeExtensionAPI])
    const activeQueryState = useObservable(
        useMemo(() => wrapRemoteObservable(sourcegraphVSCodeExtensionAPI.observeActiveWebviewQueryState()), [
            sourcegraphVSCodeExtensionAPI,
        ])
    )
    const dynamicFilters =
        useObservable(
            useMemo(() => wrapRemoteObservable(sourcegraphVSCodeExtensionAPI.observeActiveWebviewDynamicFilters()), [
                sourcegraphVSCodeExtensionAPI,
            ])
        ) ?? undefined
    const settingsCascade = useObservable(
        useMemo(() => wrapRemoteObservable(sourcegraphVSCodeExtensionAPI.getSettings()), [
            sourcegraphVSCodeExtensionAPI,
        ])
    ) ?? { final: {}, subjects: [] }

    useEffect(() => {
        // On changes that originate from user input in the search webview panel itself,
        // we don't want to trigger another query state update, which would lead to an infinite loop.
        // That's why we set the state directly, instead of using the `setQueryState` method which
        // updates query state in the search webview panel.
        if (activeQueryState) {
            useQueryState.setState({ queryState: activeQueryState.queryState })
            setPatternType(activeQueryState.patternType)
            setCaseSensitive(activeQueryState.caseSensitive)
            setActiveSearchPanel(activeQueryState.executed)
        }
    }, [activeQueryState, sourcegraphVSCodeExtensionAPI, useQueryState])

    // On submit, validate access token and update VS Code settings through API.
    // Open search tab on successful validation.
    const onSubmitAccessToken: React.FormEventHandler<HTMLFormElement> = event => {
        event?.preventDefault()
        setValidating(true)
        ;(async () => {
            const newAccessToken = (event.currentTarget.elements.namedItem('token') as HTMLInputElement).value

            if (newAccessToken) {
                await sourcegraphVSCodeExtensionAPI.updateAccessToken(newAccessToken)
                // Updating below states  would call useEffect to validate the updated token
                setHasAccessToken(true)
                setHasAccount(true)
            }
        })().catch(error => {
            console.error(error)
        })
        setValidating(false)
    }

    // There's no ACTIVE search panel

    // We need to add API to query all open search panels

    // If no open, show button + CTA to open search panel (links to sign up etc.)
    if (
        !validating &&
        onDesktop !== undefined &&
        hasAccessToken !== undefined &&
        authenticatedUser !== undefined &&
        localRecentSearches !== undefined
    ) {
        if (activeSearchPanel) {
            return (
                <BrandedSearchSidebar
                    forceButton={true}
                    className={styles.sidebarContainer}
                    filters={dynamicFilters}
                    useQueryState={useQueryState}
                    patternType={patternType}
                    caseSensitive={caseSensitive}
                    settingsCascade={settingsCascade}
                    telemetryService={{
                        log: () => {},
                        logViewEvent: () => {},
                    }}
                />
            )
        }
        return authenticatedUser || localRecentSearches ? (
            <>
                <HistorySidebar
                    sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                    platformContext={platformContext}
                    theme={theme}
                    authenticatedUser={authenticatedUser}
                    patternType={patternType}
                    caseSensitive={caseSensitive}
                    useQueryState={useQueryState}
                    localRecentSearches={localRecentSearches}
                />
            </>
        ) : (
            <>
                <OpenSearchPanelCta
                    sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                    onDesktop={onDesktop}
                />
                <SidebarAuthCheck
                    sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                    hasAccessToken={hasAccessToken}
                    telemetryService={platformContext.telemetryService}
                    onSubmitAccessToken={onSubmitAccessToken}
                    validAccessToken={authenticatedUser !== null}
                />
            </>
        )
    }
    return <LoadingSpinner />
}
