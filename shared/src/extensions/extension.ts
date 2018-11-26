import * as GQL from '../graphql/schema'
import { Settings } from '../settings/settings'
import { ErrorLike, isErrorLike } from '../util/errors'
import { ExtensionManifest, parseExtensionManifestOrError } from './extensionManifest'

/**
 * Describes a configured extension.
 */
export interface ConfiguredExtension {
    /**
     * The extension's extension ID.
     *
     * @example "alice/myextension"
     */
    readonly id: string

    /** The parsed extension manifest, null if there is none, or a parse error. */
    readonly manifest: ExtensionManifest | null | ErrorLike

    /** The raw extension manifest (JSON), or null if there is none. */
    readonly rawManifest: string | null
}

/**
 * Describes a configured extension with an optional associated registry extension. Prefer using
 * {@link ConfiguredExtension} when it is not necessary to access the registry extension's metadata.
 *
 * @template X the registry extension type
 */
export interface ConfiguredRegistryExtension<
    X extends Pick<GQL.IRegistryExtension, 'id' | 'url' | 'viewerCanAdminister'> = Pick<
        GQL.IRegistryExtension,
        'id' | 'url' | 'viewerCanAdminister'
    >
> extends ConfiguredExtension {
    /** The extension's metadata on the registry, if this is a registry extension. */
    readonly registryExtension?: X
}

type MinimalRegistryExtension = Pick<GQL.IRegistryExtension, 'extensionID' | 'id' | 'url' | 'viewerCanAdminister'> & {
    manifest: { raw: string } | null
}

/**
 * Converts to a {@link ConfiguredRegistryExtension} value.
 *
 * @template X the extension type
 */
export function toConfiguredRegistryExtension<X extends MinimalRegistryExtension>(
    extension: X
): ConfiguredRegistryExtension<X> {
    return {
        id: extension.extensionID,
        manifest: extension.manifest ? parseExtensionManifestOrError(extension.manifest.raw) : null,
        rawManifest: (extension && extension.manifest && extension.manifest.raw) || null,
        registryExtension: extension,
    }
}

/** Reports whether the given extension is enabled in the settings. */
export function isExtensionEnabled(settings: Settings | ErrorLike | null, extensionID: string): boolean {
    return !!settings && !isErrorLike(settings) && !!settings.extensions && !!settings.extensions[extensionID]
}
