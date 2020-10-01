// We want to polyfill first.
import '../../shared/polyfills'

import * as React from 'react'
import { render } from 'react-dom'
import { from, noop, Observable, Subscription, combineLatest } from 'rxjs'
import { GraphQLResult } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { background } from '../web-extension-api/runtime'
import { observeStorageKey, storage } from '../web-extension-api/storage'
import { OptionsContainer, OptionsContainerProps } from '../options-page/OptionsContainer'
import { OptionsMenuProps } from '../options-page/OptionsMenu'
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
import { map } from 'rxjs/operators'
import { isExtension } from '../../shared/context'

assertEnvironment('OPTIONS')

initSentry('options')

const IS_EXTENSION = true

interface State {
    sourcegraphURL: string | null
    isActivated: boolean
    optionFlags: OptionFlagWithValue[]
}

const isOptionFlagKey = (key: string): key is OptionFlagKey =>
    !!optionFlagDefinitions.find(definition => definition.key === key)

const fetchCurrentTabStatus = async (): Promise<{OptionsMenuProps['currentTabStatus']}> => {
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
function requestGraphQL<T, V = object>(options: { request: string; variables: V; sourcegraphURL?: string }): Observable<GraphQLResult<T>> {
    return from(background.requestGraphQL<T, V>(options))
}

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

class Options extends React.Component<{}, State> {
    public state: State = {
        sourcegraphURL: null,
        isActivated: true,
        optionFlags: [],
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            observeOptionFlagsWithValues().subscribe(optionFlags => {
                this.setState({ optionFlags })
            })
        )

        this.subscriptions.add(
            observeSourcegraphURL(IS_EXTENSION).subscribe(sourcegraphURL => {
                this.setState({ sourcegraphURL })
            })
        )

        this.subscriptions.add(
            observeStorageKey('sync', 'disableExtension').subscribe(disableExtension => {
                this.setState({
                    isActivated: !disableExtension,
                })
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        if (this.state.sourcegraphURL === null) {
            return null
        }

        const props: OptionsContainerProps = {
            sourcegraphURL: this.state.sourcegraphURL,
            isActivated: this.state.isActivated,

            ensureValidSite,
            fetchCurrentTabStatus,
            hasPermissions: url =>
                browser.permissions.contains({
                    origins: [`${url}/*`],
                }),
            requestPermissions: url =>
                browser.permissions.request({
                    origins: [`${url}/*`],
                }),

            setSourcegraphURL: (sourcegraphURL: string) => storage.sync.set({ sourcegraphURL }),
            toggleExtensionDisabled: (isActivated: boolean) => storage.sync.set({ disableExtension: !isActivated }),
            onChangeOptionFlag: (key: string, value: boolean) => {
                if (isOptionFlagKey(key)) {
                    featureFlags.set(key, value).then(noop, noop)
                }
            },
            optionFlags: this.state.optionFlags,
        }

        return <OptionsContainer {...props} />
    }
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
