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
