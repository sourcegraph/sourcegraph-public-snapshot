import { Contributions } from 'cxp/lib/protocol'
import * as GQL from '../backend/graphqlschema'
import { SourcegraphExtension } from '../schema/extension.schema'
import { parseJSON } from '../settings/configuration'
import { asError, ErrorLike } from '../util/errors'

/**
 * Describes a configured extension and its contributions. This value is propagated throughout the application.
 */
export interface ConfiguredExtension extends Pick<GQL.IConfiguredExtension, 'extensionID'> {
    /** The extension's contributions. */
    contributions?: Contributions

    /** The merged settings for the extension for the viewer. */
    settings: ExtensionSettings

    /** The parsed extension manifest, null if there is none, or a parse error. */
    manifest: SourcegraphExtension | null | ErrorLike
}

/** The settings for an extension (from global, organization, and user settings). */
export interface ExtensionSettings {
    merged: any
}

/** An extension result from the GraphQL API. */
export interface RawConfiguredExtension
    extends Pick<GQL.IConfiguredExtension, 'extensionID' | 'contributions' | 'mergedSettings'> {
    rawManifest: string | null
}

/** Parses a RawExtension into an Extension. */
export function fromRawExtension(raw: RawConfiguredExtension): ConfiguredExtension {
    let manifest: SourcegraphExtension | null | ErrorLike
    try {
        manifest = raw.rawManifest ? parseJSON(raw.rawManifest) : null
    } catch (err) {
        manifest = asError(err)
    }
    return {
        extensionID: raw.extensionID,
        contributions: raw.contributions,
        settings: { merged: raw.mergedSettings },
        manifest,
    }
}
