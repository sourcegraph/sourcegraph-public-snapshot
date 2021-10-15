import * as Comlink from 'comlink'
import { BehaviorSubject, of } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import vscode from 'vscode'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { ProxySubscribable, proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'
import { QueryState } from '@sourcegraph/shared/src/search/helpers'
import { Filter } from '@sourcegraph/shared/src/search/stream'

import { QueryStateWithInputProps, SourcegraphVSCodeSearchWebviewAPI } from '../contract'

export interface SearchSidebarMediator {
    addSearchWebviewPanel: (
        webviewPanel: vscode.WebviewPanel,
        sourcegraphVSCodeSearchWebviewAPI: Comlink.Remote<SourcegraphVSCodeSearchWebviewAPI>
    ) => void

    observeActiveWebviewQueryState: () => ProxySubscribable<QueryStateWithInputProps | null>
    observeActiveWebviewDynamicFilters: () => ProxySubscribable<Filter[] | null>
    setActiveWebviewQueryState: (queryState: QueryState) => Promise<void>
    submitActiveWebviewSearch: (queryState?: QueryState) => Promise<void>
}

/**
 * TODO: explain why
 *
 * Theoretically, a search webview could have been initialized before the search sidebar.
 *
 * Runs in extension, not webview.
 */
export function createSearchSidebarMediator(disposables: vscode.Disposable[]): SearchSidebarMediator {
    const activeSearchWebviewPanel = new BehaviorSubject<{
        webviewPanel: vscode.WebviewPanel
        sourcegraphVSCodeSearchWebviewAPI: Comlink.Remote<SourcegraphVSCodeSearchWebviewAPI>
    } | null>(null)

    // First panel + sidebar initialization order isn't deterministic, so wait
    // for both to be initialized to reveal the sidebar.
    const subscription = activeSearchWebviewPanel.subscribe(searchPanel => {
        if (searchPanel) {
            vscode.commands.executeCommand('sourcegraph.searchSidebar.focus').then(
                () => {},
                error => {
                    console.error(error)
                }
            )
        }
    })

    disposables.push({ dispose: () => subscription.unsubscribe() })

    return {
        addSearchWebviewPanel: (webviewPanel, sourcegraphVSCodeSearchWebviewAPI) => {
            if (webviewPanel.active) {
                // Make it active
                activeSearchWebviewPanel.next({
                    webviewPanel,
                    sourcegraphVSCodeSearchWebviewAPI,
                })
            }

            webviewPanel.onDidChangeViewState(event => {
                if (event.webviewPanel.active) {
                    // Make it active
                    activeSearchWebviewPanel.next({
                        webviewPanel,
                        sourcegraphVSCodeSearchWebviewAPI,
                    })
                } else if (activeSearchWebviewPanel.value?.webviewPanel === webviewPanel) {
                    // Null it out. Was previously the active webview panel
                    activeSearchWebviewPanel.next(null)
                }
            }, disposables)

            webviewPanel.onDidDispose(() => {
                // If this was the active webview panel, null it out.
                if (activeSearchWebviewPanel.value?.webviewPanel === webviewPanel) {
                    activeSearchWebviewPanel.next(null)
                }
            }, disposables)
        },

        // Methods exposed for search sidebar -> search webview (will be called for the active search webview panel's )
        observeActiveWebviewQueryState: () =>
            // Allows search sidebar to observe query state of active search webview.
            // RxJS + Comlink magic to expose an Obsverable across two message channels (search webview -> VSC extension host -> search sidebar)
            proxySubscribable(
                activeSearchWebviewPanel.pipe(
                    switchMap(searchWebview => {
                        if (!searchWebview) {
                            return of(null)
                        }

                        const { sourcegraphVSCodeSearchWebviewAPI } = searchWebview
                        return wrapRemoteObservable(sourcegraphVSCodeSearchWebviewAPI.observeQueryState())
                    })
                )
            ),
        observeActiveWebviewDynamicFilters: () =>
            proxySubscribable(
                activeSearchWebviewPanel.pipe(
                    switchMap(searchWebview => {
                        if (!searchWebview) {
                            return of(null)
                        }

                        const { sourcegraphVSCodeSearchWebviewAPI } = searchWebview
                        return wrapRemoteObservable(sourcegraphVSCodeSearchWebviewAPI.observeDynamicFilters())
                    })
                )
            ),
        setActiveWebviewQueryState: async queryState => {
            if (activeSearchWebviewPanel.value) {
                try {
                    const { sourcegraphVSCodeSearchWebviewAPI } = activeSearchWebviewPanel.value

                    // TODO decide where to do error handling. Feels like fire and forget (w/ error logging) is ok here.
                    await sourcegraphVSCodeSearchWebviewAPI.setQueryState(queryState)
                } catch (error) {
                    console.error(error)
                }
            }
        },
        submitActiveWebviewSearch: async queryState => {
            if (activeSearchWebviewPanel.value) {
                try {
                    const { sourcegraphVSCodeSearchWebviewAPI } = activeSearchWebviewPanel.value

                    await sourcegraphVSCodeSearchWebviewAPI.submitSearch(queryState)
                } catch (error) {
                    console.error(error)
                }
            }
        },
    }
}
