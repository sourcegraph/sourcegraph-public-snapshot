import '../platform/polyfills'

import * as Comlink from 'comlink'
import React from 'react'
import { render } from 'react-dom'
import { Observable } from 'rxjs'

import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import { AnchorLink, setLinkComponent } from '@sourcegraph/shared/src/components/Link'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { QueryStateWithInputProps, SourcegraphVSCodeExtensionAPI, SourcegraphVSCodeSearchWebviewAPI } from '../contract'
import { createPlatformContext, WebviewPageProps } from '../platform/context'
import { createEndpointsForWebToNode } from '../platform/webviewEndpoint'
import { adaptToEditorTheme } from '../theme'

import { SearchPage } from './SearchPage'
import { createUseQueryState, State } from './state'

const vsCodeApi = window.acquireVsCodeApi<State['state']>()

export const useQueryState = createUseQueryState(vsCodeApi)

const webviewAPI: SourcegraphVSCodeSearchWebviewAPI = {
    observeQueryState: () => {
        const queryStates = new Observable<QueryStateWithInputProps>(subscriber => {
            const cleanup = useQueryState.subscribe<QueryStateWithInputProps>(
                queryState => {
                    subscriber.next(queryState)
                },
                ({ state }) => ({
                    queryState: state.queryState,
                    caseSensitive: state.caseSensitive,
                    patternType: state.patternType,
                })
            )
            // Initial state
            const initialState = useQueryState.getState().state

            subscriber.next({
                queryState: initialState.queryState,
                caseSensitive: initialState.caseSensitive,
                patternType: initialState.patternType,
            })

            return () => {
                cleanup()
            }
        })

        return proxySubscribable(queryStates)
    },
    setQueryState: queryState => {
        useQueryState.getState().actions.setQuery(queryState)
    },
    submitSearch: queryState => {
        useQueryState.getState().actions.submitQuery(queryState)
    },
    observeDynamicFilters: () => {
        const dynamicFilters = new Observable<Filter[] | null>(subscriber => {
            const cleanup = useQueryState.subscribe<Filter[] | null>(
                filters => {
                    subscriber.next(filters)
                },
                ({ state }) => state.searchResults?.search?.results.dynamicFilters ?? null
            )

            const initialState = useQueryState.getState().state

            subscriber.next(initialState.searchResults?.search?.results.dynamicFilters ?? null)

            return () => {
                cleanup()
            }
        })

        return proxySubscribable(dynamicFilters)
    },
}

const { proxy, expose } = createEndpointsForWebToNode(vsCodeApi)

Comlink.expose(webviewAPI, expose)

const sourcegraphVSCodeExtensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI> = Comlink.wrap(proxy)

sourcegraphVSCodeExtensionAPI.panelInitialized(document.documentElement.dataset.panelId!).catch(() => {
    // TODO
})

const platformContext = createPlatformContext(sourcegraphVSCodeExtensionAPI)

setLinkComponent(AnchorLink)

const themes = adaptToEditorTheme()

const Main: React.FC = () => {
    const theme = useObservable(themes) || 'theme-dark'

    const commonPageProps: WebviewPageProps = {
        sourcegraphVSCodeExtensionAPI,
        platformContext,
        theme,
    }

    return <SearchPage {...commonPageProps} />
}
render(<Main />, document.querySelector('#root'))
