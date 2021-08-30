import { once } from 'lodash'
import { combineLatest, from, Observable, of } from 'rxjs'
import { fromFetch } from 'rxjs/fetch'
import { catchError, distinctUntilChanged, map, publishReplay, refCount, switchMap } from 'rxjs/operators'

import { checkOk } from '../../backend/fetch'
import { ConfiguredExtension, isExtensionEnabled } from '../../extensions/extension'
import { ExtensionManifest } from '../../extensions/extensionManifest'
import { areExtensionsSame } from '../../extensions/extensions'
import { viewerConfiguredExtensions } from '../../extensions/helpers'
import { PlatformContext } from '../../platform/context'
import { isErrorLike } from '../../util/errors'

/**
 * The manifest of an extension sideloaded during local development.
 *
 * Doesn't include {@link ExtensionManifest#url}, as this is added when
 * publishing an extension to the registry.
 * Instead, the bundle URL is computed from the manifest's `main` field.
 */
interface SideloadedExtensionManifest extends Omit<ExtensionManifest, 'url'> {
    name: string
    main: string
}

export const getConfiguredSideloadedExtension = (baseUrl: string): Observable<ConfiguredExtension> =>
    fromFetch(`${baseUrl}/package.json`, { selector: response => checkOk(response).json() }).pipe(
        map(
            (response: SideloadedExtensionManifest): ConfiguredExtension => ({
                id: response.name,
                manifest: {
                    ...response,
                    url: `${baseUrl}/${response.main.replace('dist/', '')}`,
                },
                rawManifest: null,
            })
        )
    )

/**
 * Returns an Observable of extensions enabled for the user.
 * Wrapped with the `once` function from lodash.
 */
export const getEnabledExtensions = once(
    (
        context: Pick<
            PlatformContext,
            'settings' | 'requestGraphQL' | 'sideloadedExtensionURL' | 'getScriptURLForExtension'
        >
    ): Observable<ConfiguredExtension[]> => {
        const sideloadedExtension: Observable<ConfiguredExtension | null> = from(context.sideloadedExtensionURL).pipe(
            switchMap(url => (url ? getConfiguredSideloadedExtension(url) : of(null))),
            catchError(error => {
                console.error('Error sideloading extension', error)
                return of(null)
            })
        )

        return combineLatest([viewerConfiguredExtensions(context), sideloadedExtension, context.settings]).pipe(
            map(([configuredExtensions, sideloadedExtension, settings]) => {
                let enabled = configuredExtensions.filter(extension => isExtensionEnabled(settings.final, extension.id))
                if (sideloadedExtension) {
                    if (!isErrorLike(sideloadedExtension.manifest) && sideloadedExtension.manifest?.publisher) {
                        // Disable extension with the same ID while this extension is sideloaded
                        const constructedID = `${sideloadedExtension.manifest.publisher}/${sideloadedExtension.id}`
                        enabled = enabled.filter(extension => extension.id !== constructedID)
                    }

                    enabled.push(sideloadedExtension)
                }
                return enabled
            }),
            distinctUntilChanged((a, b) => areExtensionsSame(a, b)),
            publishReplay(1),
            refCount()
        )
    }
)
