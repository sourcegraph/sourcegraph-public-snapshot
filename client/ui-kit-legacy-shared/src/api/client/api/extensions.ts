import { Remote } from 'comlink'
import { from, Subscription } from 'rxjs'
import { bufferCount, startWith, withLatestFrom } from 'rxjs/operators'
import { splitExtensionID } from '../../../extensions/extension'
import { getEnabledExtensionsForSubject } from '../../../extensions/extensions'
import { PlatformContext } from '../../../platform/context'
import { hashCode } from '../../../util/hashCode'
import { ExtensionExtensionsAPI } from '../../extension/api/extensions'
import { ExecutableExtension, IExtensionsService } from '../services/extensionsService'

/** @internal */
export class ClientExtensions {
    private subscriptions = new Subscription()

    /**
     * Implements the client side of the extensions API.
     *
     * @param proxy The connection to the extension host.
     * @param extensions An observable that emits the set of extensions that should be activated
     * upon subscription and whenever it changes.
     * @param platformContext The platform context
     */
    constructor(
        private proxy: Remote<ExtensionExtensionsAPI>,
        extensionRegistry: IExtensionsService,
        private platformContext: Pick<PlatformContext, 'telemetryService' | 'settings'>
    ) {
        this.subscriptions.add(
            from(extensionRegistry.activeExtensions)
                .pipe(
                    startWith([] as ExecutableExtension[]),
                    bufferCount(2, 1),
                    withLatestFrom(platformContext.settings)
                )
                .subscribe(([[oldExtensions, newExtensions], settings]) => {
                    // Diff next state's activated extensions vs. current state's.
                    if (!newExtensions) {
                        newExtensions = oldExtensions
                    }
                    const toActivate = [...newExtensions] // clone to avoid mutating state stored by bufferCount
                    const toDeactivate: ExecutableExtension[] = []
                    const next: ExecutableExtension[] = []
                    if (oldExtensions) {
                        for (const extension of oldExtensions) {
                            const newIndex = toActivate.findIndex(({ id }) => extension.id === id)
                            if (newIndex === -1) {
                                // Extension is no longer activated
                                toDeactivate.push(extension)
                            } else {
                                // Extension is already activated.
                                toActivate.splice(newIndex, 1)
                                next.push(extension)
                            }
                        }
                    }

                    /**
                     * Deactivate extensions that are no longer in use. In practice,
                     * {@link activeExtensions} never deactivates extensions, so this will never be
                     * called (in the current implementation).
                     */
                    for (const extension of toDeactivate) {
                        this.proxy.$deactivateExtension(extension.id).catch(error => {
                            console.warn(`Error deactivating extension ${JSON.stringify(extension.id)}:`, error)
                        })
                    }

                    // Activate extensions that haven't yet been activated.
                    for (const extension of toActivate) {
                        console.log('Activating Sourcegraph extension:', extension.id)

                        // We could log the event after the activation promise resolves to ensure that there wasn't
                        // an error during activation, but we want to track the maximum number of times an extension could have been useful.
                        // Since extension activation is passive from the user's perspective, and we don't yet track extension usage events,
                        // there's no way that we could measure how often extensions are actually useful anyways.

                        const defaultExtensions = getEnabledExtensionsForSubject(settings, 'DefaultSettings') || {}

                        // We only want to log non-default extension events
                        if (!defaultExtensions[extension.id]) {
                            // Hash extension IDs that specify host, since that means that it's a private registry extension.
                            const telemetryExtensionID = splitExtensionID(extension.id).host
                                ? hashCode(extension.id, 20)
                                : extension.id

                            this.platformContext.telemetryService?.log('ExtensionActivation', {
                                extension_id: telemetryExtensionID,
                            })
                        }

                        this.proxy.$activateExtension(extension.id, extension.scriptURL).catch(error => {
                            console.error(`Error activating extension ${JSON.stringify(extension.id)}:`, error)
                        })
                    }
                })
        )
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
