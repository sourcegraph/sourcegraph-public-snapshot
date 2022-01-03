import classNames from 'classnames'
import React, { useEffect, useMemo, useState } from 'react'
import { Form } from 'reactstrap'
import create, { UseStore } from 'zustand'

import { SearchSidebar as BrandedSearchSidebar } from '@sourcegraph/branded/src/search/results/sidebar/SearchSidebar'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import { CurrentAuthStateResult, CurrentAuthStateVariables } from '@sourcegraph/shared/src/graphql-operations'
import { SearchQueryState, updateQuery } from '@sourcegraph/shared/src/search/searchQueryState'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { WebviewPageProps } from '../platform/context'

import { OpenSearchPanelCta } from './OpenSearchPanelCta'
import styles from './SearchSidebar.module.scss'
import { SidebarAuthCheck } from './SidebarAuthCheck'

interface SearchSidebarProps extends Pick<WebviewPageProps, 'platformContext' | 'sourcegraphVSCodeExtensionAPI'> {}

export const SearchSidebar: React.FC<SearchSidebarProps> = ({ sourcegraphVSCodeExtensionAPI, platformContext }) => {
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
            // useQueryState.getState().setQueryState(activeQueryState)
            useQueryState.setState({ queryState: activeQueryState.queryState })
        }
    }, [activeQueryState, useQueryState])

    // Check if User is currently on VS Code Desktop or VS Code Web
    const [onDesktop, setOnDesktop] = useState<boolean | undefined>(undefined)
    const [hasAccessToken, setHasAccessToken] = useState<boolean | undefined>(undefined)
    const [corsUri, setCorsUri] = useState<string | undefined>(undefined)
    const [validating, setValidating] = useState(true)
    const [hasAccount, setHasAccount] = useState(false)
    const [validAccessToken, setValidAccessToken] = useState<boolean>(false)

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
                // TODO error handling
                .catch(() => setHasAccessToken(false))
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
                    setValidAccessToken(true)
                    await sourcegraphVSCodeExtensionAPI.openSearchPanel()
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

    const onSubmitCorsUrl: React.FormEventHandler<HTMLFormElement> = event => {
        event?.preventDefault()
        setValidating(true)
        ;(async () => {
            const newUri = (event.currentTarget.elements.namedItem('corsuri') as HTMLInputElement).value

            if (corsUri !== newUri) {
                await sourcegraphVSCodeExtensionAPI.updateCorsUri(newUri)
                // Updating below states  would call useEffect to validate the updated token
                setCorsUri(newUri)
            }
        })().catch(error => {
            console.error(error)
        })
        setValidating(false)
    }

    if (!activeQueryState?.queryState.query) {
        // There's no ACTIVE search panel

        // We need to add API to query all open search panels

        // If no open, show button + CTA to open search panel (links to sign up etc.)

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
                {!validating && !onDesktop && corsUri === '' && (
                    <Form onSubmit={onSubmitCorsUrl}>
                        <p className="btn btn-sm btn-danger w-100 border-0 font-weight-normal">
                            <span className={classNames('my-3', styles.text)}>
                                IMPORTANT: You must add Cors and have a Sourcegraph account for Sourcegraph to work on
                                VS Code Web
                            </span>
                        </p>
                        <input
                            className="input form-control my-3"
                            type="text"
                            name="corsuri"
                            placeholder="ex https://cors-anywhere.herokuapp.com/"
                        />
                        <button
                            type="submit"
                            className={classNames(
                                'btn btn-sm btn-link w-100 border-0 font-weight-normal',
                                styles.button
                            )}
                        >
                            <span className={classNames('my-0', styles.text)}>Add Cors</span>
                        </button>
                    </Form>
                )}
                {validating && <LoadingSpinner />}
            </>
        )
    }
    // For v1: Add recent/saved searches/files panel(s)
    const { caseSensitive, patternType } = activeQueryState

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
