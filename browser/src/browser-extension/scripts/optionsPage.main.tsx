// We want to polyfill first.
import '../../shared/polyfills'

import * as React from 'react'
import { render } from 'react-dom'
import { from, noop, Observable, Subscription } from 'rxjs'
import { GraphQLResult } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { background } from '../web-extension-api/runtime'
import { observeStorageKey, storage } from '../web-extension-api/storage'
import { featureFlagDefaults, FeatureFlags } from '../web-extension-api/types'
import { OptionsContainer, OptionsContainerProps } from '../options-page/OptionsContainer'
import { OptionsMenuProps } from '../options-page/OptionsMenu'
import { initSentry } from '../../shared/sentry'
import { fetchSite } from '../../shared/backend/server'
import { featureFlags } from '../../shared/util/featureFlags'
import { OptionFlagKey } from '../../shared/util/optionFlags'
import { assertEnvironment } from '../environmentAssertion'
import { observeSourcegraphURL } from '../../shared/util/context'

assertEnvironment('OPTIONS')

initSentry('options')

const IS_EXTENSION = true

type State = Pick<
    FeatureFlags,
    'allowErrorReporting' | 'experimentalLinkPreviews' | 'experimentalTextFieldCompletion' | 'sendTelemetry'
> & { sourcegraphURL: string | null; isActivated: boolean }

const keyIsFeatureFlag = (key: string): key is keyof FeatureFlags =>
    !!Object.keys(featureFlagDefaults).find(featureFlag => key === featureFlag)

const fetchCurrentTabStatus = async (): Promise<OptionsMenuProps['currentTabStatus']> => {
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
function requestGraphQL<T, V = object>(options: { request: string; variables: V }): Observable<GraphQLResult<T>> {
    return from(background.requestGraphQL<T, V>(options))
}

const observeOptionFlags = (): Observable<Partial<FeatureFlags> | undefined> =>
    observeStorageKey('sync', 'featureFlags')

const ensureValidSite = (): Observable<GQL.ISite> => fetchSite(requestGraphQL)

class Options extends React.Component<{}, State> {
    public state: State = {
        sourcegraphURL: null,
        isActivated: true,

        // Feature flags
        allowErrorReporting: false,
        sendTelemetry: false,
        experimentalLinkPreviews: false,
        experimentalTextFieldCompletion: false,
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            observeOptionFlags().subscribe(optionFlags => {
                const {
                    allowErrorReporting,
                    experimentalLinkPreviews,
                    experimentalTextFieldCompletion,
                    sendTelemetry,
                } = {
                    ...featureFlagDefaults,
                    ...optionFlags,
                }
                this.setState({
                    allowErrorReporting,
                    experimentalLinkPreviews,
                    experimentalTextFieldCompletion,
                    sendTelemetry,
                })
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
            onChangeOptionFlag: (key: OptionFlagKey, value: boolean) => {
                if (keyIsFeatureFlag(key)) {
                    featureFlags.set(key, value).then(noop, noop)
                }
            },
            optionFlags: [
                {
                    key: 'sendTelemetry',
                    label: 'Send telemetry',
                    value: this.state.sendTelemetry,
                },
                {
                    key: 'allowErrorReporting',
                    label: 'Allow error reporting',
                    value: this.state.allowErrorReporting,
                },
                {
                    key: 'experimentalLinkPreviews',
                    label: 'Experimental link previews',
                    value: this.state.experimentalLinkPreviews,
                },
                {
                    key: 'experimentalTextFieldCompletion',
                    label: 'Experimental text field completion',
                    value: this.state.experimentalTextFieldCompletion,
                },
            ],
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
