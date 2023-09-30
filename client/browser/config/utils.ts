/**
 * Generates a unique bundle ID that is used to prevent the Phabricator extension
 * from returning cached contents after upgrading.
 *
 * @returns The current extension version from extension.info.json.
 */
export function generateBundleUID(): string {
    // Static version from `client/browser/src/browser-extension/manifest.spec.json`
    // TODO: make it dynamic and figure out if we need to use manifest.spec.json here.
    return '0.0.0'
}
