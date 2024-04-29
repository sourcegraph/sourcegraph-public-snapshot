/**
 * Returns whether or not to use the embedded Cody Pro subscription management UI,
 * or direct users to https://accounts.sourcegraph.com.
 */
export function showEmbeddedCodyProUi(): boolean {
    return window?.context?.frontendCodyProConfig?.useEmbeddedUi === true
}
