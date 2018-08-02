/**
 * A subset of the settings JSON Schema type containing the minimum needed by this library.
 */
export interface Settings {
    extensions?: { [extensionID: string]: boolean }
    [key: string]: any
}
