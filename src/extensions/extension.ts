import { Extension } from 'cxp/module/environment/extension'
import { Settings } from '../copypasta'
import { ErrorLike } from '../errors'
import { SourcegraphExtension } from '../schema/extension.schema'
import * as GQL from '../schema/graphqlschema'

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
    manifest: SourcegraphExtension | null | ErrorLike

    /** The raw extension manifest (JSON), or null if there is none.. */
    rawManifest: string | null

    /** The corresponding extension on the registry, if any. */
    registryExtension?: RX
}

/** Reports whether the given extension is enabled in the settings. */
export function isExtensionEnabled(settings: Settings | null, extensionID: string): boolean {
    return !!settings && !!settings.extensions && !!settings.extensions[extensionID]
}

/** Reports whether the given extension is mentioned (enabled or disabled) in the settings. */
export function isExtensionAdded(settings: Settings | null, extensionID: string): boolean {
    return !!settings && !!settings.extensions && extensionID in settings.extensions
}
