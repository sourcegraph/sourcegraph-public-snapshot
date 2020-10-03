// We want to polyfill first.
import '../../shared/polyfills'

import * as React from 'react'
import { render } from 'react-dom'
import { from, noop, Observable, Subscription, combineLatest } from 'rxjs'
import { GraphQLResult } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
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
import { observeSourcegraphURL, isFirefox } from '../../shared/util/context'
import { catchError, map, mapTo, tap } from 'rxjs/operators'
import { isExtension } from '../../shared/context'
import { OptionsPage } from '../options-menu/OptionsPage'
import { asError } from '../../../../shared/src/util/errors'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { AnchorLink, setLinkComponent } from '../../../../shared/src/components/Link'
import { useMemo } from 'react'

assertEnvironment('OPTIONS')

initSentry('options')

const IS_EXTENSION = true

setLinkComponent(AnchorLink)

interface State {
    sourcegraphURL: string | null
    isActivated: boolean
    optionFlags: OptionFlagWithValue[]
}

const isOptionFlagKey = (key: string): key is OptionFlagKey =>
    !!optionFlagDefinitions.find(definition => definition.key === key)

const fetchCurrentTabStatus = async (): Promise<{ host: string; protocol: string; hasPermissions: boolean }> => {
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

const isFullPage = (): boolean => !new URLSearchParams(window.location.search).get('popup')

const validateSourcegraphUrl = (url: string): Observable<string | undefined> =>
    fetchSite(options => requestGraphQL({ ...options, sourcegraphURL: url })).pipe(
        tap(value => console.log('Response', { url, value })),
        mapTo(undefined),
        catchError(error => asError(error).message)
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

const ensureValidSite = (): Observable<GQL.ISite> => fetchSite(requestGraphQL)

const Options: React.FunctionComponent = () => {
    const sourcegraphUrl = useObservable(useMemo(() => observeSourcegraphURL(true), [])) || ''

    return (
        <OptionsPage
            isFullPage={isFullPage()}
            isCurrentRepositoryPrivate={false} // TODO
            sourcegraphUrl={sourcegraphUrl}
            version="dev"
            isActivated={true}
            validateSourcegraphUrl={validateSourcegraphUrl}
            onToggleActivated={noop} // TODO
        />
    )
}

const inject = (): void => {
    const injectDOM = document.createElement('div')
    injectDOM.className = 'sourcegraph-options-menu options'
    document.body.append(injectDOM)
    // For shared CSS that would otherwise be dark by default
    document.body.classList.add('theme-light')

    render(<Options />, injectDOM)
}

document.addEventListener('DOMContentLoaded', inject)
