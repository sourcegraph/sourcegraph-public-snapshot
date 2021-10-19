import * as Comlink from 'comlink'
import React from 'react'
import { render } from 'react-dom'
import { BehaviorSubject } from 'rxjs'

import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { SourcegraphVSCodeExtensionAPI, SourcegraphVSCodeWebviewAPI } from './contract'
import { SearchPage } from './search'

declare global {
    // eslint-disable-next-line no-var
    var acquireVsCodeApi: () => {
        postMessage: (message: any) => void
    }
}

// TODO: can probably use zustand for all state + serialization?
const routeSubject = new BehaviorSubject<'search' | undefined>(undefined)

const webviewAPI: SourcegraphVSCodeWebviewAPI = {
    setRoute: route => {
        routeSubject.next(route)
    },
}

const vscodeApi = window.acquireVsCodeApi()

/**
 * TODO explain (using `type` since we can't achieve bi-directional Comlink connections by transferring MessageChannels
 * like we do in the web app or browser extension)
 */
function vscodeWebviewEndpoint(connectionType: 'webview' | 'extension'): Comlink.Endpoint {
    const onMessages = new WeakMap<EventListenerOrEventListenerObject, EventListener>()

    return {
        postMessage: message => vscodeApi.postMessage({ ...message, connectionType }),
        addEventListener: (type, listener) => {
            // Listener will be used as key for wrapped listener which filters for connectionType
            function onMessage(event: MessageEvent): void {
                if (event.data?.connectionType === connectionType) {
                    return typeof listener === 'function' ? listener(event) : listener.handleEvent(event)
                }
            }

            onMessages.set(listener, onMessage as EventListener)

            window.addEventListener('message', onMessage)
        },
        removeEventListener: (type, listener) => {
            const onMessage = onMessages.get(listener)

            window.removeEventListener('message', onMessage ?? listener)
        },
    }
}

Comlink.expose(webviewAPI, vscodeWebviewEndpoint('webview'))

const sourcegraphVSCodeExtensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI> = Comlink.wrap(
    vscodeWebviewEndpoint('extension')
)

console.log({ sourcegraphVSCodeExtensionAPI })

sourcegraphVSCodeExtensionAPI
    .ping()
    .then(value => console.log({ value }))
    .catch(() => {
        // noop
    })

export interface WebviewPageProps {
    sourcegraphVSCodeExtensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI>
}

const Main: React.FC = () => {
    const route = useObservable(routeSubject)

    const commonPageProps: WebviewPageProps = {
        sourcegraphVSCodeExtensionAPI,
    }

    if (route === undefined) {
        return <p>Loading...</p>
    }

    return <SearchPage {...commonPageProps} />
}
render(<Main />, document.querySelector('#root'))
