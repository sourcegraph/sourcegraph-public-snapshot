// We want to polyfill first.
import '../../shared/polyfills'

import React, { useEffect, useState } from 'react'
import { render } from 'react-dom'
import { from, noop, Observable, combineLatest } from 'rxjs'
import { GraphQLResult } from '../../../../shared/src/graphql/graphql'
import { background } from '../web-extension-api/runtime'
import { observeStorageKey, storage } from '../web-extension-api/storage'
import { initSentry } from '../../shared/sentry'
import { fetchSite } from '../../shared/backend/server'
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
import { observeSourcegraphURL, isFirefox, getExtensionVersion } from '../../shared/util/context'
import { catchError, map, mapTo } from 'rxjs/operators'
import { isExtension } from '../../shared/context'
import { OptionsPage } from '../options-menu/OptionsPage'
import { asError } from '../../../../shared/src/util/errors'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { AnchorLink, setLinkComponent } from '../../../../shared/src/components/Link'
import { KnownCodeHost, knownCodeHosts } from '../knownCodeHosts'
import { Optional } from 'utility-types'

interface TabStatus {
    host: string
    protocol: string
    hasPermissions: boolean
}

assertEnvironment('OPTIONS')

initSentry('options')

const IS_EXTENSION = true

/**
 * A list of protocols where we should *not* show the permissions notification.
 */
const PERMISSIONS_PROTOCOL_BLOCKLIST = new Set(['chrome:', 'about:'])

setLinkComponent(AnchorLink)

const isOptionFlagKey = (key: string): key is OptionFlagKey =>
    !!optionFlagDefinitions.find(definition => definition.key === key)

const fetchCurrentTabStatus = async (): Promise<TabStatus> => {
    const tabs = await browser.tabs.query({ active: true, currentWindow: true })
    if (tabs.length > 1) {
        throw new Error('Querying for the currently active tab returned more than one result')
    }
    const { url } = tabs[0]
    if (!url) {
        throw new Error('Currently active tab has no URL')
    }
    const { host, protocol } = new URL(url)
    const hasPermissions = await browser.permissions.contains({
        origins: [`${protocol}//${host}/*`],
    })
    return { host, protocol, hasPermissions }
}

// Make GraphQL requests from background page
function requestGraphQL<T, V = object>(options: {
    request: string
    variables: V
    sourcegraphURL?: string
}): Observable<GraphQLResult<T>> {
    return from(background.requestGraphQL<T, V>(options))
}

const version = getExtensionVersion()
const isFullPage = !new URLSearchParams(window.location.search).get('popup')

const validateSourcegraphUrl = (url: string): Observable<string | undefined> =>
    fetchSite(options => requestGraphQL({ ...options, sourcegraphURL: url })).pipe(
        mapTo(undefined),
        catchError(error => [asError(error).message])
    )

const observeOptionFlagsWithValues = (): Observable<OptionFlagWithValue[]> => {
    const overrideSendTelemetry: Observable<boolean> = observeSourcegraphURL(IS_EXTENSION).pipe(
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
const observingSourcegraphUrl = observeSourcegraphURL(true)
const observingOptionFlagsWithValues = observeOptionFlagsWithValues()

function handleToggleActivated(isActivated: boolean): void {
    storage.sync.set({ disableExtension: !isActivated }).catch(console.error)
}

function handleChangeOptionFlag(key: string, value: boolean): void {
    if (isOptionFlagKey(key)) {
        featureFlags.set(key, value).catch(noop)
    }
}

function buildRequestPermissionsHandler({ protocol, host }: TabStatus) {
    return function requestPermissionsHandler(event: React.MouseEvent) {
        event.preventDefault()
        browser.permissions.request({ origins: [`${protocol}//${host}/*`] }).catch(noop)
    }
}

const Options: React.FunctionComponent = () => {
    const sourcegraphUrl = useObservable(observingSourcegraphUrl) || ''
    const isActivated = useObservable(observingIsActivated)
    const optionFlagsWithValues = useObservable(observingOptionFlagsWithValues) || []
    const [currentTabStatus, setCurrentTabStatus] = useState<
        { status: TabStatus; handler: React.MouseEventHandler } | undefined
    >()

    useEffect(() => {
        fetchCurrentTabStatus().then(tabStatus => {
            setCurrentTabStatus({ status: tabStatus, handler: buildRequestPermissionsHandler(tabStatus) })
        }, noop)
    }, [])

    let permissionAlert: Optional<KnownCodeHost, 'host' | 'icon'> | undefined
    if (
        currentTabStatus &&
        !currentTabStatus?.status.hasPermissions &&
        !PERMISSIONS_PROTOCOL_BLOCKLIST.has(currentTabStatus.status.protocol)
    ) {
        const knownCodeHost = knownCodeHosts.find(({ host }) => host === currentTabStatus.status.host)
        if (knownCodeHost) {
            permissionAlert = knownCodeHost
        } else {
            permissionAlert = { name: currentTabStatus.status.host }
        }
    }

    /**
     * TODO(tj): Finish permissions logic, then implement private repo logic
     * - Observe permissions (browser.permissions.onAdded), set currentTabStatus in subscription
     */

    return (
        <OptionsPage
            isFullPage={isFullPage}
            sourcegraphUrl={sourcegraphUrl}
            version={version}
            validateSourcegraphUrl={validateSourcegraphUrl}
            isActivated={!!isActivated}
            onToggleActivated={handleToggleActivated}
            optionFlags={optionFlagsWithValues}
            onChangeOptionFlag={handleChangeOptionFlag}
            showPrivateRepositoryAlert={false}
            permissionAlert={permissionAlert}
            currentHost={currentTabStatus?.status.host}
            requestPermissionsHandler={currentTabStatus?.handler}
        />
    )
}

const inject = (): void => {
    document.body.classList.add('theme-light')
    render(<Options />, document.body)
}

document.addEventListener('DOMContentLoaded', inject)
