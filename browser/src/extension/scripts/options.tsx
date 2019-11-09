// We want to polyfill first.
import '../polyfills'

import * as React from 'react'
import { render } from 'react-dom'
import { from, noop, Observable } from 'rxjs'
import { GraphQLResult } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { background } from '../../browser/runtime'
import { observeStorageKey, storage } from '../../browser/storage'
import { featureFlagDefaults, FeatureFlags } from '../../browser/types'
import { OptionsMenu, CurrentTabStatus } from '../../libs/options/OptionsMenu'
import { initSentry } from '../../libs/sentry'
import { fetchSite } from '../../shared/backend/server'
import { featureFlags } from '../../shared/util/featureFlags'
import { assertEnv } from '../envAssertion'
import { getExtensionVersion } from '../../shared/util/context'
import { filter, map, mapTo } from 'rxjs/operators'
import { isDefined } from '../../../../shared/src/util/types'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { useSourcegraphURL } from '../../libs/options/useSourcegraphURL'

assertEnv('OPTIONS')

initSentry('options')

const keyIsFeatureFlag = (key: string): key is keyof FeatureFlags =>
    !!Object.keys(featureFlagDefaults).find(k => key === k)

const toggleFeatureFlag = (key: string): void => {
    if (keyIsFeatureFlag(key)) {
        featureFlags
            .toggle(key)
            .then(noop)
            .catch(noop)
    }
}

const fetchCurrentTabStatus = async (): Promise<CurrentTabStatus> => {
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

const connectToSourcegraphInstance = (baseURL: string): Observable<void> => {
    const requestGraphQL = <T extends GQL.IQuery | GQL.IMutation>({
        request,
        variables,
    }: {
        request: string
        variables: {}
    }): Observable<GraphQLResult<T>> =>
        from(
            background.requestGraphQL<T>({ request, variables, baseURL })
        )
    return fetchSite(requestGraphQL).pipe(mapTo(undefined))
}

const persistSourcegraphURL = (sourcegraphURL: string): Observable<void> => from(storage.sync.set({ sourcegraphURL }))

const onToggleActivationClick = (isActivated: boolean): void => {
    storage.sync
        .set({ disableExtension: !isActivated })
        .catch(err => console.error('Error setting disableExtension', err))
}

const observeSourcegraphURL = (): Observable<string> =>
    observeStorageKey('sync', 'sourcegraphURL').pipe(filter(isDefined))

const urlHasPermissions = (url: string): Observable<boolean> =>
    from(
        browser.permissions.contains({
            origins: [`${url}/*`],
        })
    )

const requestPermissions = (url: string): void => {
    browser.permissions
        .request({
            origins: [`${url}/*`],
        })
        .catch(err => console.error(`Error requesting permissions for ${url}`, err))
}

const Options: React.FunctionComponent = () => {
    const featureFlags = useObservable(
        React.useMemo(
            () =>
                observeStorageKey('sync', 'featureFlags').pipe(
                    map(featureFlags => {
                        const { allowErrorReporting, experimentalLinkPreviews, experimentalTextFieldCompletion } = {
                            ...featureFlagDefaults,
                            ...featureFlags,
                        }
                        return {
                            allowErrorReporting,
                            experimentalLinkPreviews,
                            experimentalTextFieldCompletion,
                        }
                    })
                ),
            []
        )
    )
    const isActivated =
        useObservable(
            React.useMemo(
                () => observeStorageKey('sync', 'disableExtension').pipe(map(disableExtension => !disableExtension)),
                []
            )
        ) ?? true
    const version = getExtensionVersion()
    const [onSourcegraphURLChange, onSourcegraphURLSubmit, sourcegraphURLAndStatus] = useSourcegraphURL({
        connectToSourcegraphInstance,
        persistSourcegraphURL,
        urlHasPermissions,
        observeSourcegraphURL,
    })
    const [currentTabStatus, setCurrentTabStatus] = React.useState<CurrentTabStatus>()
    React.useEffect(() => {
        fetchCurrentTabStatus()
            .then(status => setCurrentTabStatus(status))
            .catch(err => console.error('Error fetching current tab status', err))
    }, [])
    if (!sourcegraphURLAndStatus) {
        return null
    }
    const { sourcegraphURL, connectionStatus } = sourcegraphURLAndStatus
    return (
        <OptionsMenu
            toggleFeatureFlag={toggleFeatureFlag}
            requestPermissions={requestPermissions}
            featureFlags={featureFlags}
            version={version}
            isActivated={isActivated}
            currentTabStatus={currentTabStatus}
            onToggleActivationClick={onToggleActivationClick}
            onSourcegraphURLChange={onSourcegraphURLChange}
            onSourcegraphURLSubmit={onSourcegraphURLSubmit}
            sourcegraphURL={sourcegraphURL}
            connectionStatus={connectionStatus}
        />
    )
}

const inject = (): void => {
    const injectDOM = document.createElement('div')
    injectDOM.className = 'sourcegraph-options-menu options'
    document.body.appendChild(injectDOM)
    // For shared CSS that would otherwise be dark by default
    document.body.classList.add('theme-light')

    render(<Options />, injectDOM)
}

document.addEventListener('DOMContentLoaded', inject)
