import * as GQL from '../backend/graphqlschema'
import { Contributions } from './contributions'

/**
 * Describes a configured extension and its contributions. This value is propagated throughout the application.
 */
export interface Extension extends Pick<GQL.IConfiguredExtension, 'extensionID'> {
    /** The extension's contributions. */
    contributions?: Contributions

    /** The merged settings for the extension for the viewer. */
    settings: ExtensionSettings
}

/** The settings for an extension (from global, organization, and user settings). */
export interface ExtensionSettings {
    merged: any
}

/** An extension result from the GraphQL API. */
export interface RawExtension
    extends Pick<GQL.IConfiguredExtension, 'extensionID' | 'contributions' | 'mergedSettings'> {}

/** Parses a RawExtension into an Extension. */
export function fromRawExtension(raw: RawExtension): Extension {
    return {
        extensionID: raw.extensionID,
        contributions: raw.contributions,
        settings: { merged: raw.mergedSettings },
    }
}
