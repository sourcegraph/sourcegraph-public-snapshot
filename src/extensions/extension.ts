import { Settings } from '../copypasta'
import { ErrorLike } from '../errors'
import { SourcegraphExtension } from '../schema/extension.schema'
import * as GQL from '../schema/graphqlschema'
import { ConfigurationSubject, ConfiguredSubject } from '../settings'

/**
 * Describes a configured extension.
 *
 * @template S the configuration subject type
 * @template C the type of the extension's settings (overlaid on the base settings JSON Schema-derived type)
 * @template RX the registry extension type
 */
export interface ConfiguredExtension<
    S extends ConfigurationSubject = ConfigurationSubject,
    C extends { [key: string]: any } = Settings,
    RX extends Pick<GQL.IRegistryExtension, 'id' | 'url' | 'viewerCanAdminister'> = Pick<
        GQL.IRegistryExtension,
        'id' | 'url' | 'viewerCanAdminister'
    >
> {
    /** The ID of the extension, unique on a Sourcegraph site. */
    extensionID: string

    /** The merged settings for the extension for the viewer. */
    settings: C | null // TODO(sqs): make this interface extend ConfigurationCascade<ExtensionSettings>
    // TODO(sqs): make this also | ErrorLike maybe? unless we remove this field altogether

    /** The settings for the extension at each level of the cascade. */
    settingsCascade: ConfiguredSubject<S, C>[]

    /** Whether the extension is enabled for the viewer. */
    isEnabled: boolean

    /** Whether the extension is added in the viewer's settings. */
    isAdded: boolean

    /** The parsed extension manifest, null if there is none, or a parse error. */
    manifest: SourcegraphExtension | null | ErrorLike

    /** The raw extension manifest (JSON), or null if there is none.. */
    rawManifest: string | null

    /** The corresponding extension on the registry, if any. */
    registryExtension?: RX
}

export function isExtensionEnabled(settings: Settings | null, extensionID: string): boolean {
    return !!settings && !!settings.extensions && !!settings.extensions[extensionID]
}

export function isExtensionAdded(settings: Settings | null, extensionID: string): boolean {
    return !!settings && !!settings.extensions && extensionID in settings.extensions
}
