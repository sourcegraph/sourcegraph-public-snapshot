import classNames from 'classnames'
import React, { useEffect, useMemo, useState } from 'react'
import { Form } from 'reactstrap'
import create, { UseStore } from 'zustand'

import { SearchSidebar as BrandedSearchSidebar } from '@sourcegraph/branded/src/search/results/sidebar/SearchSidebar'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { AuthenticatedUser, currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import {
    CurrentAuthStateResult,
    CurrentAuthStateVariables,
    SearchPatternType,
} from '@sourcegraph/shared/src/graphql-operations'
import { SearchQueryState, updateQuery } from '@sourcegraph/shared/src/search/searchQueryState'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { LocalRecentSeachProps } from '../contract'
import { WebviewPageProps } from '../platform/context'

import { OpenSearchPanelCta } from './OpenSearchPanelCta'
import { SearchHistoryPanel } from './SearchHistoryPanel'
import styles from './SearchSidebar.module.scss'
import { SidebarAuthCheck } from './SidebarAuthCheck'
interface SearchSidebarProps extends WebviewPageProps {}

export const SearchSidebar: React.FC<SearchSidebarProps> = ({
    sourcegraphVSCodeExtensionAPI,
    platformContext,
    theme,
}) => {
    // Check if there is any opened / active search panel
    const [activeSearchPanel, setActiveSearchPanel] = useState<boolean>(false)
    const [patternType, setPatternType] = useState<SearchPatternType>(SearchPatternType.literal)
    const [caseSensitive, setCaseSensitive] = useState<boolean>(false)
    const useQueryState: UseStore<SearchQueryState> = useMemo(() => {
        const useStore = create<SearchQueryState>((set, get) => ({
            queryState: { query: '' },
            setQueryState: queryStateUpdate => {
                const queryState =
                    typeof queryStateUpdate === 'function' ? queryStateUpdate(get().queryState) : queryStateUpdate
                set({ queryState })
                // TODO error handling

                sourcegraphVSCodeExtensionAPI.setActiveWebviewQueryState(queryState).then(
                    () => {},
                    () => {}
                )
            },
            submitSearch: (_parameters, updates = []) => {
                const updatedQuery = updateQuery(get().queryState.query, updates)
                // TODO error handling
                sourcegraphVSCodeExtensionAPI
                    .submitActiveWebviewSearch({
                        query: updatedQuery,
                    })
                    .then(
                        () => {},
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
            if (activeQueryState.executed) {
                setActiveSearchPanel(true)
                sourcegraphVSCodeExtensionAPI
                    .getLocalRecentSearch()
                    .then(response => {
                        setLocalRecentSearches(response)
                    })
                    .catch(() => {
                        // TODO error handling
                    })
            } else {
                setActiveSearchPanel(false)
            }
        }
    }, [activeQueryState, sourcegraphVSCodeExtensionAPI, useQueryState])

    // Check if User is currently on VS Code Desktop or VS Code Web
    const [onDesktop, setOnDesktop] = useState<boolean | undefined>(undefined)
    const [hasAccessToken, setHasAccessToken] = useState<boolean | undefined>(undefined)
    const [corsUri, setCorsUri] = useState<string | undefined>(undefined)
    const [validating, setValidating] = useState(true)
    const [hasAccount, setHasAccount] = useState(false)
    const [validAccessToken, setValidAccessToken] = useState<boolean>(false)
    const [localRecentSearches, setLocalRecentSearches] = useState<LocalRecentSeachProps[] | undefined>(undefined)
    const [authenticatedUser, setAuthenticatedUser] = useState<AuthenticatedUser | null>(null)

    // Get current access token, cros, and platform settings
    useEffect(() => {
        setValidating(true)
        // Get initial settings
        if (onDesktop === undefined) {
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
            // Get Local Search History
            sourcegraphVSCodeExtensionAPI
                .getLocalRecentSearch()
                .then(response => {
                    setLocalRecentSearches(response)
                })
                .catch(() => {
                    // TODO error handling
                })
        }
        // Validate Token
        if (!validAccessToken) {
            ;(async () => {
                const currentUser = await platformContext
                    .requestGraphQL<CurrentAuthStateResult, CurrentAuthStateVariables>({
                        request: currentAuthStateQuery,
                        variables: {},
                        mightContainPrivateInfo: true,
                    })
                    .toPromise()
                if (currentUser.data) {
                    // If user is detected, set valid access token to true
                    setAuthenticatedUser(currentUser.data.currentUser)
                    setValidAccessToken(true)
                } else {
                    setValidAccessToken(false)
                }
            })().catch(error => console.error(error))
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
        localRecentSearches,
    ])

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

    // const onSubmitCorsUrl: React.FormEventHandler<HTMLFormElement> = event => {
    //     event?.preventDefault()
    //     setValidating(true)
    //     ;(async () => {
    //         const newUri = (event.currentTarget.elements.namedItem('corsuri') as HTMLInputElement).value

    //         if (corsUri !== newUri) {
    //             await sourcegraphVSCodeExtensionAPI.updateCorsUri(newUri)
    //             // Updating below states  would call useEffect to validate the updated token
    //             setCorsUri(newUri)
    //         }
    //     })().catch(error => {
    //         console.error(error)
    //     })
    //     setValidating(false)
    // }

    // There's no ACTIVE search panel

    // We need to add API to query all open search panels

    // If no open, show button + CTA to open search panel (links to sign up etc.)
    if (!activeSearchPanel && !validAccessToken) {
        return (
            <>
                {!validating && onDesktop !== undefined && (
                    <OpenSearchPanelCta
                        className={styles.sidebarContainer}
                        sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                        onDesktop={onDesktop}
                    />
                )}
                {!validating && hasAccessToken !== undefined && (
                    <SidebarAuthCheck
                        className={styles.sidebarContainer}
                        sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                        hasAccessToken={hasAccessToken}
                        telemetryService={platformContext.telemetryService}
                        onSubmitAccessToken={onSubmitAccessToken}
                        validAccessToken={validAccessToken}
                    />
                )}
                {/* If User is not on VS Code Desktop and do not have Cors set up */}
                {!validating && !onDesktop && !hasAccessToken && (
                    <Form onSubmit={onSubmitAccessToken}>
                        <p className="btn btn-sm btn-danger w-100 border-0 font-weight-normal">
                            <span className={classNames('my-3', styles.text)}>
                                IMPORTANT: You must add an access token for Sourcegraph to work on VS Code Web
                            </span>
                        </p>
                        <input
                            className="input form-control my-3"
                            type="text"
                            name="token"
                            placeholder="ex 6dfc880b320dff712d9f6cfcac5cbd13ebfad1d8"
                        />
                        <button
                            type="submit"
                            className={classNames(
                                'btn btn-sm btn-link w-100 border-0 font-weight-normal',
                                styles.button
                            )}
                        >
                            <span className={classNames('my-0', styles.text)}>Add Access Token</span>
                        </button>
                    </Form>
                )}
                {validating && <LoadingSpinner />}
            </>
        )
    }

    // For v1: Add recent/saved searches/files panel(s)

    return (
        <>
            {/* HISTORY SIDEBAR */}
            <SearchHistoryPanel
                localRecentSearches={localRecentSearches}
                telemetryService={platformContext.telemetryService}
                authenticatedUser={authenticatedUser}
                platformContext={platformContext}
                sourcegraphVSCodeExtensionAPI={sourcegraphVSCodeExtensionAPI}
                theme={theme}
            />
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
        </>
    )
}
