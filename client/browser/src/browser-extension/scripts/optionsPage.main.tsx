// We want to polyfill first.
import '../../shared/polyfills'

import React, { useEffect, useState } from 'react'
import { render } from 'react-dom'
import { from, noop, Observable, combineLatest } from 'rxjs'
import { catchError, map, mapTo } from 'rxjs/operators'
import { Optional } from 'utility-types'

import { AnchorLink, setLinkComponent } from '@sourcegraph/shared/src/components/Link'
import { GraphQLResult } from '@sourcegraph/shared/src/graphql/graphql'
import { isFirefox } from '@sourcegraph/shared/src/util/browserDetection'
import { asError } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { fetchSite } from '../../shared/backend/server'
import { isExtension } from '../../shared/context'
import { SourcegraphUrlService } from '../../shared/platform/sourcegraphUrlService'
import { initSentry } from '../../shared/sentry'
import { getExtensionVersion, CLOUD_SOURCEGRAPH_URL } from '../../shared/util/context'
import { featureFlags } from '../../shared/util/featureFlags'
import {
    OptionFlagKey,
    OptionFlagWithValue,
    assignOptionFlagValues,
    observeOptionFlags,
    shouldOverrideSendTelemetry,
    optionFlagDefinitions,
} from '../../shared/util/optionFlags'
import { assertEnvironment } from '../environmentAssertion'
import { KnownCodeHost, knownCodeHosts } from '../knownCodeHosts'
import { URL_AUTH_ERROR, URL_FETCH_ERROR } from '../options-menu/constants'
import { OptionsPage } from '../options-menu/OptionsPage'
import { OptionsPageContext } from '../options-menu/OptionsPage.context'
import { ThemeWrapper } from '../ThemeWrapper'
import { background } from '../web-extension-api/runtime'
import { observeStorageKey, storage } from '../web-extension-api/storage'

interface TabStatus {
    host: string
    protocol: string
    hasPermissions: boolean
    hasPrivateCloudError: boolean
}

assertEnvironment('OPTIONS')

initSentry('options')

/**
 * A list of protocols where we should *not* show the permissions notification.
 */
const PERMISSIONS_PROTOCOL_BLOCKLIST = new Set(['chrome:', 'about:', 'safari-web-extension:'])

setLinkComponent(AnchorLink)

const isOptionFlagKey = (key: string): key is OptionFlagKey =>
    !!optionFlagDefinitions.find(definition => definition.key === key)

const fetchCurrentTabStatus = async (): Promise<TabStatus> => {
    const tabs = await browser.tabs.query({ active: true, currentWindow: true })
    if (tabs.length > 1) {
        throw new Error('Querying for the currently active tab returned more than one result')
    }
    const { url, id } = tabs[0]
    if (!url) {
        throw new Error('Currently active tab has no URL')
    }
    if (!id) {
        throw new Error('Currently active tab has no ID')
    }
    const hasPrivateCloudError = await background.checkPrivateCloudError(id)
    const { host, protocol } = new URL(url)
    const hasPermissions = await browser.permissions.contains({
        origins: [`${protocol}//${host}/*`],
    })
    return { hasPrivateCloudError, host, protocol, hasPermissions }
}

// Make GraphQL requests from background page
function requestGraphQL<T, V = object>(options: {
    request: string
    variables: V
    sourcegraphURL: string
}): Observable<GraphQLResult<T>> {
    return from(background.requestGraphQL<T, V>(options))
}

const version = getExtensionVersion()
const isFullPage = !new URLSearchParams(window.location.search).get('popup')

const validateSourcegraphUrl = (sourcegraphURL: string): Observable<string | undefined> =>
    fetchSite(options => requestGraphQL({ ...options, sourcegraphURL })).pipe(
        mapTo(undefined),
        catchError(error => {
            const { message } = asError(error)
            // We lose Error type when communicating from the background page
            // to the options page, so we determine the error type from the message
            if (message.includes('Failed to fetch')) {
                return [URL_FETCH_ERROR]
            }
            if (message.includes('401')) {
                return [URL_AUTH_ERROR]
            }

            return [message]
        })
    )

const observeOptionFlagsWithValues = (): Observable<OptionFlagWithValue[]> => {
    const overrideSendTelemetry: Observable<boolean> = SourcegraphUrlService.observe().pipe(
        map(sourcegraphUrl => shouldOverrideSendTelemetry(isFirefox(), isExtension, sourcegraphUrl))
    )

    return combineLatest([observeOptionFlags(), overrideSendTelemetry]).pipe(
        map(([flags, override]) => {
            const definitions = assignOptionFlagValues(flags)
            if (override) {
                return definitions.filter(flag => flag.key !== 'sendTelemetry')
            }
            return definitions
        })
    )
}

const observingIsActivated = observeStorageKey('sync', 'disableExtension').pipe(map(isDisabled => !isDisabled))
const observingOptionFlagsWithValues = observeOptionFlagsWithValues()

function handleToggleActivated(isActivated: boolean): void {
    storage.sync.set({ disableExtension: !isActivated }).catch(console.error)
}

function onChangeOptionFlag(key: string, value: boolean): void {
    if (isOptionFlagKey(key)) {
        featureFlags.set(key, value).catch(noop)
    }
}

function handleSelfHostedSourcegraphURLChange(sourcegraphURL?: string): void {
    SourcegraphUrlService.setSelfHostedSourcegraphURL(sourcegraphURL).catch(console.error)
}

function onBlocklistChange(enabled: boolean, content: string): void {
    SourcegraphUrlService.setBlocklist({ enabled, content }).catch(console.error)
}

function buildRequestPermissionsHandler({ protocol, host }: TabStatus) {
    return function requestPermissionsHandler(event: React.MouseEvent) {
        event.preventDefault()
        browser.permissions.request({ origins: [`${protocol}//${host}/*`] }).catch(error => {
            console.error('Error requesting permissions:', error)
        })
    }
}

const Options: React.FunctionComponent = () => {
    const sourcegraphURL = useObservable(SourcegraphUrlService.observe())
    const selfHostedSourcegraphURL = useObservable(SourcegraphUrlService.getSelfHostedSourcegraphURL())
    const blocklist = useObservable(SourcegraphUrlService.getBlocklist())
    const isActivated = useObservable(observingIsActivated)
    const optionFlags = useObservable(observingOptionFlagsWithValues) || []
    const [currentTabStatus, setCurrentTabStatus] = useState<
        { status: TabStatus; handler: React.MouseEventHandler } | undefined
    >()

    useEffect(() => {
        fetchCurrentTabStatus().then(tabStatus => {
            setCurrentTabStatus({ status: tabStatus, handler: buildRequestPermissionsHandler(tabStatus) })
        }, noop)
    }, [])

    const showSourcegraphCloudAlert = currentTabStatus?.status.host.endsWith('sourcegraph.com')

    let permissionAlert: Optional<KnownCodeHost, 'host' | 'icon'> | undefined
    if (
        currentTabStatus &&
        !currentTabStatus?.status.hasPermissions &&
        !showSourcegraphCloudAlert &&
        !PERMISSIONS_PROTOCOL_BLOCKLIST.has(currentTabStatus.status.protocol)
    ) {
        const knownCodeHost = knownCodeHosts.find(({ host }) => host === currentTabStatus.status.host)
        if (knownCodeHost) {
            permissionAlert = knownCodeHost
        } else {
            permissionAlert = { name: currentTabStatus.status.host }
        }
    }

    return (
        <ThemeWrapper>
            <OptionsPageContext.Provider
                value={{
                    blocklist,
                    onBlocklistChange,
                    optionFlags,
                    onChangeOptionFlag,
                }}
            >
                <OptionsPage
                    isFullPage={isFullPage}
                    selfHostedSourcegraphURL={selfHostedSourcegraphURL}
                    version={version}
                    validateSourcegraphUrl={validateSourcegraphUrl}
                    permissionAlert={permissionAlert}
                    onSelfHostedSourcegraphURLChange={handleSelfHostedSourcegraphURLChange}
                    isActivated={!!isActivated}
                    onToggleActivated={handleToggleActivated}
                    showPrivateRepositoryAlert={
                        currentTabStatus?.status.hasPrivateCloudError && sourcegraphURL === CLOUD_SOURCEGRAPH_URL
                    }
                    showSourcegraphCloudAlert={showSourcegraphCloudAlert}
                    requestPermissionsHandler={currentTabStatus?.handler}
                />
            </OptionsPageContext.Provider>
        </ThemeWrapper>
    )
}

const inject = (): void => {
    render(<Options />, document.body)
}

document.addEventListener('DOMContentLoaded', inject)
