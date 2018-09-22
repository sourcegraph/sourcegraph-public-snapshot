import { Extension } from 'sourcegraph/module/client/extension'
import { ErrorLike, isErrorLike } from '../errors'
import { ExtensionManifest } from '../schema/extension.schema'
import * as GQL from '../schema/graphqlschema'
import { Settings } from '../settings'

/**
 * Describes a configured extension.
 *
 * @template S the configuration subject type
 * @template C the type of the extension's settings (overlaid on the base settings JSON Schema-derived type)
 * @template RX the registry extension type
 */
export interface ConfiguredExtension<
    RX extends Pick<GQL.IRegistryExtension, 'id' | 'url' | 'viewerCanAdminister'> = Pick<
        GQL.IRegistryExtension,
        'id' | 'url' | 'viewerCanAdminister'
    >
> extends Extension {
    /** The parsed extension manifest, null if there is none, or a parse error. */
    manifest: ExtensionManifest | null | ErrorLike

    /** The raw extension manifest (JSON), or null if there is none. */
    rawManifest: string | null

    /** The corresponding extension on the registry, if any. */
    registryExtension?: RX
}

/** Reports whether the given extension is enabled in the settings. */
export function isExtensionEnabled(settings: Settings | ErrorLike | null, extensionID: string): boolean {
    return !!settings && !isErrorLike(settings) && !!settings.extensions && !!settings.extensions[extensionID]
}

/** Reports whether the given extension is mentioned (enabled or disabled) in the settings. */
export function isExtensionAdded(settings: Settings | ErrorLike | null, extensionID: string): boolean {
    return !!settings && !isErrorLike(settings) && !!settings.extensions && extensionID in settings.extensions
}

/**
 * Shows a modal confirmation prompt to the user confirming whether to add an extension.
 */
export function confirmAddExtension(extensionID: string, extensionManifest?: ConfiguredExtension['manifest']): boolean {
    // Either `"title" (id)` (if there is a title in the manifest) or else just `id`. It is
    // important to show the ID because it indicates who the publisher is and allows
    // disambiguation from other similarly titled extensions.
    let displayName: string
    if (extensionManifest && !isErrorLike(extensionManifest) && extensionManifest.title) {
        displayName = `${JSON.stringify(extensionManifest.title)} (${extensionID})`
    } else {
        displayName = extensionID
    }
    return confirm(
        `Add Sourcegraph extension ${displayName}?\n\nIt can:\n- Read repositories and files you view using Sourcegraph\n- Read and change your Sourcegraph settings`
    )
}
