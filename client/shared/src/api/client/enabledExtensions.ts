import { isEqual, once } from 'lodash'
import { combineLatest, from, Observable, of, throwError } from 'rxjs'
import { fromFetch } from 'rxjs/fetch'
import { catchError, distinctUntilChanged, map, publishReplay, refCount, shareReplay, switchMap } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { checkOk } from '@sourcegraph/http-client'

import {
    ConfiguredExtension,
    ConfiguredExtensionManifestDefaultFields,
    extensionIDsFromSettings,
    isExtensionEnabled,
} from '../../extensions/extension'
import { ExtensionManifest } from '../../extensions/extensionManifest'
import { areExtensionsSame } from '../../extensions/extensions'
import { queryConfiguredRegistryExtensions } from '../../extensions/helpers'
import { PlatformContext } from '../../platform/context'
import { isSettingsValid } from '../../settings/settings'

/**
 * @returns An observable that emits the list of extensions configured in the viewer's final settings upon
 * subscription and each time it changes.
 */
function viewerConfiguredExtensions({
    settings,
    getGraphQLClient,
}: Pick<PlatformContext, 'settings' | 'getGraphQLClient'>): Observable<ConfiguredExtension[]> {
    return from(settings).pipe(
        map(settings => extensionIDsFromSettings(settings)),
        distinctUntilChanged((a, b) => isEqual(a, b)),
        switchMap(extensionIDs => queryConfiguredRegistryExtensions({ getGraphQLClient }, extensionIDs)),
        catchError(error => throwError(asError(error))),
        // TODO: Restore reference counter after refactoring contributions service
        // to not unsubscribe from existing entries when new entries are registered,
        // in order to ensure that the source is unsubscribed from.
        shareReplay(1)
    )
}

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

export const getConfiguredSideloadedExtension = (
    baseUrl: string
): Observable<ConfiguredExtension<ConfiguredExtensionManifestDefaultFields | 'publisher'>> =>
    fromFetch(`${baseUrl}/package.json`, { selector: response => checkOk(response).json() }).pipe(
        map(
            (response: SideloadedExtensionManifest): ConfiguredExtension => ({
                id: response.name,
                manifest: {
                    ...response,
                    url: `${baseUrl}/${response.main.replace('dist/', '')}`,
                },
            })
        )
    )

/**
 * List of extensions migrated to the core workflow. These extensions should not be activated if
 * `extensionsAsCoreFeatures` experimental feature is enabled.
 */
export const MIGRATED_TO_CORE_WORKFLOW_EXTENSION_IDS = new Set([
    'sourcegraph/git-extras',
    'sourcegraph/search-export',
    'sourcegraph/open-in-editor',
    'sourcegraph/open-in-vscode',
    'dymka/open-in-webstorm',
    'sourcegraph/open-in-atom',
])

/**
 * Returns an Observable of extensions enabled for the user.
 * Wrapped with the `once` function from lodash.
 */
export const getEnabledExtensions = once(
    (
        context: Pick<
            PlatformContext,
            | 'settings'
            | 'getGraphQLClient'
            | 'sideloadedExtensionURL'
            | 'getScriptURLForExtension'
            | 'clientApplication'
        >
    ): Observable<ConfiguredExtension[]> => {
        const sideloadedExtension = from(context.sideloadedExtensionURL).pipe(
            switchMap(url => (url ? getConfiguredSideloadedExtension(url) : of(null))),
            catchError(error => {
                console.error('Error sideloading extension', error)
                return of(null)
            })
        )

        return combineLatest([viewerConfiguredExtensions(context), sideloadedExtension, context.settings]).pipe(
            map(([configuredExtensions, sideloadedExtension, settings]) => {
                const extensionsAsCoreFeatures =
                    isSettingsValid(settings) && settings.final.experimentalFeatures?.extensionsAsCoreFeatures
                const enableGoImportsSearchQueryTransform =
                    isSettingsValid(settings) &&
                    settings.final.experimentalFeatures?.enableGoImportsSearchQueryTransform

                let enabled = configuredExtensions.filter(extension => {
                    const extensionsAsCoreFeatureMigratedExtension =
                        extensionsAsCoreFeatures && MIGRATED_TO_CORE_WORKFLOW_EXTENSION_IDS.has(extension.id)
                    // Ignore extensions migrated to the core workflow if the experimental feature is enabled
                    if (context.clientApplication === 'sourcegraph' && extensionsAsCoreFeatureMigratedExtension) {
                        return false
                    }

                    // Go import search query transform is enabled by default but can be disabled by the setting
                    const enableGoImportsSearchQueryTransformMigratedExtension =
                        (enableGoImportsSearchQueryTransform === undefined || enableGoImportsSearchQueryTransform) &&
                        extension.id === 'go-imports-search'
                    // Ignore loading the go-imports-search extension when the migrated go imports search is enabled
                    if (
                        context.clientApplication === 'sourcegraph' &&
                        enableGoImportsSearchQueryTransformMigratedExtension
                    ) {
                        return false
                    }

                    return isExtensionEnabled(settings.final, extension.id)
                })
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
