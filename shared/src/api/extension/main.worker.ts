// The ponyfill symbol-observable is impure. Since extensions are loaded through importScripts,
// if one of our extensions depends on symbol-observable, it may break other extensions:
// https://github.com/sourcegraph/sourcegraph/issues/1243
// Importing symbol-observable when starting the web worker fixes this by
// ensuring that `Symbol.observable` is mutated happens before any extensions are loaded.
import 'symbol-observable'
import { createWebWorkerMessageTransports } from '../protocol/jsonrpc2/transports/webWorker'
import { startExtensionHost } from './extensionHost'

/**
 * The entrypoint for the JavaScript context that runs the extension host (and all extensions).
 *
 * To initialize the extension host, the parent sends it an "initialize" message with
 * {@link InitData}.
 */
function extensionHostMain(): void {
    try {
        const { unsubscribe } = startExtensionHost(createWebWorkerMessageTransports())
        self.addEventListener('unload', () => unsubscribe())
    } catch (err) {
        console.error('Error starting the extension host:', err)
        self.close()
    }
}

extensionHostMain()
