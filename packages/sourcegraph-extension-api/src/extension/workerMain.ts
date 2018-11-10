import { createExtensionHost, InitData } from './extensionHost'

interface MessageEvent {
    data: any
    origin: string | null
}

/**
 * This is a subset of DedicatedWorkerGlobalScope. We can't use `/// <references lib="webworker"/>` because
 * Prettier does not support triple-slash directive syntax.
 */
interface DedicatedWorkerGlobalScope {
    location: {
        origin: string
    }
    addEventListener(type: 'message', listener: (event: MessageEvent) => void): void
    removeEventListener(type: 'message', listener: (event: MessageEvent) => void): void
    importScripts(url: string): void
    close(): void
}

/**
 * The entrypoint for Web Workers that are spawned to run an extension.
 *
 * To initialize the worker, the parent sends it a message whose data is an object conforming to the
 * {@link InitData} interface. Among other things, this contains the URL of the extension's JavaScript bundle.
 *
 * @param self The worker's `self` global scope.
 */
export function extensionHostWorkerMain(self: DedicatedWorkerGlobalScope): void {
    self.addEventListener('message', receiveExtensionURL)

    async function receiveExtensionURL(ev: MessageEvent): Promise<void> {
        try {
            // Only listen for the 1st URL.
            self.removeEventListener('message', receiveExtensionURL)

            if (ev.origin && ev.origin !== self.location.origin) {
                console.error(`Invalid extension host message origin: ${ev.origin} (expected ${self.location.origin})`)
                self.close()
            }

            const initData: InitData = ev.data
            if (typeof initData.bundleURL !== 'string') {
                console.error(`Invalid extension bundle URL: ${initData.bundleURL}`)
                self.close()
            }

            const api = createExtensionHost(initData)
            // Make `import 'sourcegraph'` or `require('sourcegraph')` return the extension host's
            // implementation of the `sourcegraph` module.
            ;(self as any).require = (modulePath: string): any => {
                if (modulePath === 'sourcegraph') {
                    return api
                }
                throw new Error(`require: module not found: ${modulePath}`)
            }

            // Load the extension bundle and retrieve the extension entrypoint module's exports on the global
            // `module` property.
            ;(self as any).exports = {}
            ;(self as any).module = {}
            self.importScripts(initData.bundleURL)
            const extensionExports = (self as any).module.exports
            delete (self as any).module

            if ('activate' in extensionExports) {
                await Promise.resolve(extensionExports.activate())
            } else {
                console.error(`Extension did not export an 'activate' function.`)
            }
        } catch (err) {
            console.error(
                `Extension host is terminating because the extension's activate function threw an error.`,
                err
            )
            // Calling self.close() here would kill the Web Worker, which sounds like a good idea (because it won't be used for
            // anything else given the error that occurred). However, this makes it harder to inspect log messages and
            // exceptions from this Web Worker in the browser devtools console. In Chrome, the filenames and line numbers from
            // track traces and log messages open a new window when clicked (instead of opening in the Sources tab). Keeping
            // the Web Worker alive fixes this issue. The Web Worker will be killed by the client when this extension is no
            // longer active, so the resource consumption of not killing unused Web Workers is bounded.
        }
    }
}
