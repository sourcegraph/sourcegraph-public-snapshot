import { ProxyMarked, proxyMarker } from 'comlink'
import { Subscription, Unsubscribable } from 'rxjs'
import { asError } from '../../../util/errors'
import { tryCatchPromise } from '../../util'

/** @internal */
export interface ExtensionExtensionsAPI extends ProxyMarked {
    $activateExtension(extensionID: string, bundleURL: string): Promise<void>
    $deactivateExtension(extensionID: string): Promise<void>
}

/** The WebWorker's global scope */
declare const self: any

/** @internal */
export class Extensions implements ExtensionExtensionsAPI, Unsubscribable, ProxyMarked {
    public readonly [proxyMarker] = true

    /** Extensions' deactivate functions. */
    private extensionDeactivate = new Map<string, () => void | Promise<void>>()

    /**
     * Proxy method invoked by the client to load an extension and invoke its `activate` function to start running it.
     *
     * It also sets up global hooks so that when the extension's code uses `require('sourcegraph')` and
     * `import 'sourcegraph'`, it gets the extension API handle (the value specified in
     * sourcegraph.d.ts).
     *
     * @param extensionID The extension ID of the extension to activate.
     * @param bundleURL The URL to the JavaScript source file (that exports an `activate` function) for
     * the extension.
     * @returns A promise that resolves when the extension's activation finishes (i.e., when it returns if it's synchronous, or when the promise it returns resolves if it's async).
     */
    public async $activateExtension(extensionID: string, bundleURL: string): Promise<void> {
        // Load the extension bundle and retrieve the extension entrypoint module's exports on
        // the global `module` property.
        try {
            const exports = {}
            self.exports = exports
            self.module = { exports }
            self.importScripts(bundleURL)
        } catch (error) {
            throw new Error(
                `error thrown while executing extension ${JSON.stringify(
                    extensionID
                )} bundle (in importScripts of ${bundleURL}): ${String(error)}`
            )
        }
        const extensionExports = self.module.exports
        delete self.exports
        delete self.module

        if (!('activate' in extensionExports)) {
            throw new Error(
                `extension bundle for ${JSON.stringify(
                    extensionID
                )} has no exported activate function (in ${bundleURL})`
            )
        }

        // During extension deactivation, we first call the extension's `deactivate` function and then unsubscribe
        // the Subscription passed to the `activate` function.
        const extensionSubscriptions = new Subscription()
        const extensionDeactivate =
            typeof extensionExports.deactivate === 'function' ? extensionExports.deactivate : null
        this.extensionDeactivate.set(extensionID, async () => {
            try {
                if (extensionDeactivate) {
                    await Promise.resolve(extensionDeactivate())
                }
            } finally {
                extensionSubscriptions.unsubscribe()
            }
        })

        // The behavior should be consistent for both sync and async activate functions that throw
        // errors or reject. Both cases should yield a rejected promise.
        //
        // TODO(sqs): Add timeouts to prevent long-running activate or deactivate functions from
        // significantly delaying other extensions.
        await tryCatchPromise<void>(() => extensionExports.activate({ subscriptions: extensionSubscriptions })).catch(
            error => {
                error = asError(error)
                throw Object.assign(
                    new Error(
                        `error during extension ${JSON.stringify(extensionID)} activate function: ${String(
                            error.stack || error
                        )}`
                    ),
                    { error }
                )
            }
        )
    }

    public async $deactivateExtension(extensionID: string): Promise<void> {
        const deactivate = this.extensionDeactivate.get(extensionID)
        if (deactivate) {
            this.extensionDeactivate.delete(extensionID)
            await Promise.resolve(deactivate())
        }
    }

    /**
     * Deactivates all activated extensions that have "deactivate" functions. It does not wait for
     * the deactivation to finish for an extension if its deactivate function is async. If any
     * deactivate functions throw an error or reject, the error is logged and not propagated.
     *
     * There is no guarantee that extensions' deactivate functions are called or that execution
     * continues until they are finished. The JavaScript execution context may be terminated before
     * deactivation is completed.
     */
    public unsubscribe(): void {
        for (const [extensionID, deactivate] of this.extensionDeactivate.entries()) {
            this.extensionDeactivate.delete(extensionID)
            tryCatchPromise(deactivate).catch(error => {
                console.warn(`Error deactivating extension ${JSON.stringify(extensionID)}:`, error)
            })
        }
    }
}
