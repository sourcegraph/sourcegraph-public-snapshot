import * as GQL from '../graphql/schema'
import { Settings, SettingsCascadeOrError } from '../settings/settings'
import { asError, ErrorLike, isErrorLike } from '../util/errors'

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

    /** The raw extension manifest (JSON), or null if there is none. */
    readonly rawManifest: string | null
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
        rawManifest: extension?.manifest?.raw || null,
        registryExtension: extension,
    }
}

/** Reports whether the given extension is enabled in the settings. */
export function isExtensionEnabled(settings: Settings | ErrorLike | null, extensionID: string): boolean {
    return !!settings && !isErrorLike(settings) && !!settings.extensions && !!settings.extensions[extensionID]
}

/**
 * Returns the extension's script URL from its manifest.
 *
 * @param extension The extension whose script URL to get.
 * @throws If the script URL can't be determined.
 * @returns The extension's script URL.
 */
export function getScriptURLFromExtensionManifest(extension: ConfiguredExtension): string {
    if (!extension.manifest) {
        throw new Error(`extension ${JSON.stringify(extension.id)}: no manifest found`)
    }
    if (isErrorLike(extension.manifest)) {
        throw new Error(`extension ${JSON.stringify(extension.id)}: invalid manifest: ${extension.manifest.message}`)
    }
    if (!extension.manifest.url) {
        throw new Error(`extension ${JSON.stringify(extension.id)}: no "url" property in manifest`)
    }
    return extension.manifest.url
}

/**
 * List of insight-like extension ids. These insights worked via extensions before,
 * but at the moment they work via insight built-in data-fetchers.
 */
const DEPRECATED_EXTENSION_IDS = new Set(['sourcegraph/code-stats-insights', 'sourcegraph/search-insights'])

/**
 * @throws An error if the final settings has an error.
 * @returns An array of extension IDs configured in the settings.
 */
export function extensionIDsFromSettings(settings: SettingsCascadeOrError): string[] {
    if (isErrorLike(settings.final)) {
        throw asError(settings.final)
    }
    if (!settings.final?.extensions) {
        return []
    }

    return (
        Object.keys(settings.final.extensions)
            // Filter out deprecated extensions
            .filter(extensionId => !DEPRECATED_EXTENSION_IDS.has(extensionId))
    )
}

/**
 * Mirrors `registry.SplitExtensionID` from `frontend`:
 *
 * `splitExtensionID` splits an extension ID of the form [host/]publisher/name (where [host/] is the
 * optional registry prefix), such as "alice/myextension" or
 * "sourcegraph.example.com/bob/myextension". It returns the components in an object.
 *
 * @param extensionID The extension ID (string)
 */
export function splitExtensionID(
    extensionID: string
): { publisher: string; name: string; host?: string; isSourcegraphExtension?: boolean } {
    const parts = extensionID.split('/')
    if (parts.length === 3) {
        return {
            host: parts[0],
            publisher: parts[1],
            name: parts[2],
        }
    }

    return {
        publisher: parts[0] ?? '',
        name: parts[1] ?? '',
        isSourcegraphExtension: parts[0] === 'sourcegraph',
    }
}
