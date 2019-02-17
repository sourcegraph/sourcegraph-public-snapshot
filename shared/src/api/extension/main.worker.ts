// The ponyfill symbol-observable is impure. Since extensions are loaded through importScripts,
// if one of our extensions depends on symbol-observable, it may break other extensions:
// https://github.com/sourcegraph/sourcegraph/issues/1243
// Importing symbol-observable when starting the web worker fixes this by
// ensuring that `Symbol.observable` is mutated happens before any extensions are loaded.
import { fromEvent } from 'rxjs'
import { take } from 'rxjs/operators'
import 'symbol-observable'
import { isEndpointPair } from '../../platform/context'
import { startExtensionHost } from './extensionHost'

/**
 * The entrypoint for the JavaScript context that runs the extension host (and all extensions).
 *
 * To initialize the extension host, the parent sends it an "initialize" message with
 * {@link InitData}.
 */
async function extensionHostMain(): Promise<void> {
    try {
        const event = await fromEvent<MessageEvent>(self, 'message')
            .pipe(take(1))
            .toPromise()
        if (!isEndpointPair(event.data)) {
            throw new Error('First message event in extension host worker did not contain MessagePort')
        }
        const endpoints = event.data
        const extensionHost = startExtensionHost(endpoints)
        self.addEventListener('unload', () => extensionHost.unsubscribe())
    } catch (err) {
        console.error('Error starting the extension host:', err)
        self.close()
    }
}

// tslint:disable-next-line: no-floating-promises
extensionHostMain()
