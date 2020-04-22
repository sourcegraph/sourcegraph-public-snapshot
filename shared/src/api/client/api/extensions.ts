import { ProxyResult } from '@sourcegraph/comlink'
import { from, Subscription } from 'rxjs'
import { bufferCount, startWith } from 'rxjs/operators'
import { ExtExtensionsAPI } from '../../extension/api/extensions'
import { ExecutableExtension, ExtensionsService } from '../services/extensionsService'

/** @internal */
export class ClientExtensions {
    private subscriptions = new Subscription()

    /**
     * Implements the client side of the extensions API.
     *
     * @param connection The connection to the extension host.
     * @param extensions An observable that emits the set of extensions that should be activated
     * upon subscription and whenever it changes.
     */
    constructor(private proxy: ProxyResult<ExtExtensionsAPI>, extensionRegistry: ExtensionsService) {
        this.subscriptions.add(
            from(extensionRegistry.activeExtensions)
                .pipe(startWith([] as ExecutableExtension[]), bufferCount(2, 1))
                .subscribe(([oldExtensions, newExtensions]) => {
                    // Diff next state's activated extensions vs. current state's.
                    const toActivate = [...newExtensions] // clone to avoid mutating state stored by bufferCount
                    const toDeactivate: ExecutableExtension[] = []
                    const next: ExecutableExtension[] = []
                    if (oldExtensions) {
                        for (const x of oldExtensions) {
                            const newIndex = toActivate.findIndex(({ id }) => x.id === id)
                            if (newIndex === -1) {
                                // Extension is no longer activated
                                toDeactivate.push(x)
                            } else {
                                // Extension is already activated.
                                toActivate.splice(newIndex, 1)
                                next.push(x)
                            }
                        }
                    }

                    /**
                     * Deactivate extensions that are no longer in use. In practice,
                     * {@link activeExtensions} never deactivates extensions, so this will never be
                     * called (in the current implementation).
                     */
                    for (const x of toDeactivate) {
                        this.proxy.$deactivateExtension(x.id).catch(err => {
                            console.warn(`Error deactivating extension ${JSON.stringify(x.id)}:`, err)
                        })
                    }

                    // Activate extensions that haven't yet been activated.
                    for (const x of toActivate) {
                        this.proxy.$activateExtension(x.id, x.scriptURL).catch(err => {
                            console.error(`Error activating extension ${JSON.stringify(x.id)}:`, err)
                        })
                    }
                })
        )
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
