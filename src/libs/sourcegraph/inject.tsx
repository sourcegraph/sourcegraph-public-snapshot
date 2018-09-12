import { connectAsClient } from '@sourcegraph/extensions-client-common/lib/messaging'
import storage from '../../browser/storage'
import { updateExtensionSettings } from '../../shared/backend/extensions'

export function injectSourcegraphApp(marker: HTMLElement): void {
    if (document.getElementById(marker.id)) {
        return
    }

    // Generate and insert DOM element, in case this code executes first.
    document.body.appendChild(marker)

    connectAsClient()
        .then(connection => {
            storage.observeSync('clientSettings').subscribe(settings => {
                connection.sendSettings(settings)
            })

            connection.onEditSetting(edit => updateExtensionSettings('Client', edit).toPromise())

            connection.onGetSettings(
                () =>
                    new Promise<string>(resolve => {
                        storage.getSync(storageItems => {
                            resolve(storageItems.clientSettings)
                        })
                    })
            )
        })
        .catch(error => console.error(error))

    window.addEventListener('load', () => {
        dispatchSourcegraphEvents()
    })

    if (document.readyState === 'complete' || document.readyState === 'interactive') {
        dispatchSourcegraphEvents()
    }
}

function dispatchSourcegraphEvents(): void {
    // Send custom webapp <-> extension registration event in case webapp listener is attached first.
    document.dispatchEvent(new CustomEvent<{}>('sourcegraph:browser-extension-registration'))
}
