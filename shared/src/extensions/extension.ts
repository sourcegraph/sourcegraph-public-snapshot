import { Extension } from '../api/client/extension'
import { ErrorLike, isErrorLike } from '../errors'
import * as GQL from '../graphqlschema'
import { ExtensionManifest } from '../schema/extension.schema'
import { Settings } from '../settings'
import { parseJSONCOrError } from '../util'

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

type MinimalRegistryExtension = Pick<GQL.IRegistryExtension, 'extensionID' | 'id' | 'url' | 'viewerCanAdminister'> & {
    manifest: { raw: string } | null
}

/**
 * Converts each element of an array to a {@link ConfiguredExtension} value.
 *
 * @template X the extension type
 */
export function toConfiguredExtensions<X extends MinimalRegistryExtension>(
    registryExtensions: X[]
): ConfiguredExtension<X>[] {
    const configuredExtensions: ConfiguredExtension<X>[] = []
    for (const registryExtension of registryExtensions) {
        configuredExtensions.push(toConfiguredExtension<X>(registryExtension))
    }
    return configuredExtensions
}

/**
 * Converts to a {@link ConfiguredExtension} value.
 *
 * @template X the extension type
 */
export function toConfiguredExtension<X extends MinimalRegistryExtension>(extension: X): ConfiguredExtension<X> {
    return {
        id: extension.extensionID,
        manifest: extension.manifest ? parseJSONCOrError<ExtensionManifest>(extension.manifest.raw) : null,
        rawManifest: (extension && extension.manifest && extension.manifest.raw) || null,
        registryExtension: extension,
    }
}

/** Reports whether the given extension is enabled in the settings. */
export function isExtensionEnabled(settings: Settings | ErrorLike | null, extensionID: string): boolean {
    return !!settings && !isErrorLike(settings) && !!settings.extensions && !!settings.extensions[extensionID]
}
