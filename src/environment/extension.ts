/** An extension's identifier and settings. */
export interface Extension {
    /** The extension ID. */
    readonly id: string

    /** The active settings for an extension. */
    readonly settings: ExtensionSettings
}

/** The settings for an extension. */
export interface ExtensionSettings {
    /** The final merged settings (from all sources). */
    readonly merged: any
}
