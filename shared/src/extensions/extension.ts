import { Extension } from '../api/client/extension'
import { ErrorLike, isErrorLike } from '../errors'
import * as GQL from '../graphqlschema'
import { ExtensionManifest } from '../schema/extension.schema'
import { Settings } from '../settings'

/**
 * Describes a configured extension.
 *
 * @template X the registry extension type
 */
export interface ConfiguredExtension<
    X extends Pick<GQL.IRegistryExtension, 'id' | 'url' | 'viewerCanAdminister'> = Pick<
        GQL.IRegistryExtension,
        'id' | 'url' | 'viewerCanAdminister'
    >
> extends Extension {
    /** The parsed extension manifest, null if there is none, or a parse error. */
    manifest: ExtensionManifest | null | ErrorLike

    /** The raw extension manifest (JSON), or null if there is none. */
    rawManifest: string | null

    /** The corresponding extension on the registry, if any. */
    registryExtension?: X
}

/** Reports whether the given extension is enabled in the settings. */
export function isExtensionEnabled(settings: Settings | ErrorLike | null, extensionID: string): boolean {
    return !!settings && !isErrorLike(settings) && !!settings.extensions && !!settings.extensions[extensionID]
}
