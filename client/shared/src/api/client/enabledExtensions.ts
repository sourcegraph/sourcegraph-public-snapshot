import { isEqual, once } from 'lodash'
import { combineLatest, from, Observable, throwError } from 'rxjs'
import { catchError, distinctUntilChanged, map, publishReplay, refCount, shareReplay, switchMap } from 'rxjs/operators'

import { asError } from '@sourcegraph/common'

import { ConfiguredExtension, extensionIDsFromSettings, isExtensionEnabled } from '../../extensions/extension'
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
 * List of extensions migrated to the core workflow.
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
            'settings' | 'getGraphQLClient' | 'getScriptURLForExtension' | 'clientApplication'
        >
    ): Observable<ConfiguredExtension[]> =>
        combineLatest([viewerConfiguredExtensions(context), context.settings]).pipe(
            map(([configuredExtensions, settings]) => {
                const enableGoImportsSearchQueryTransform =
                    isSettingsValid(settings) &&
                    settings.final.experimentalFeatures?.enableGoImportsSearchQueryTransform

                return configuredExtensions.filter(extension => {
                    const extensionsAsCoreFeatureMigratedExtension = MIGRATED_TO_CORE_WORKFLOW_EXTENSION_IDS.has(
                        extension.id
                    )
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
            }),
            distinctUntilChanged((a, b) => areExtensionsSame(a, b)),
            publishReplay(1),
            refCount()
        )
)
