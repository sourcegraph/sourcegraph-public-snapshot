/**
 * Checks if the user agent indicates the browser is running on a mobile device.
 *
 * Returns true if the user agent matches common mobile device patterns, false otherwise.
 */
export function isMobile(): boolean {
    return (
        typeof window !== 'undefined' &&
        window.navigator.userAgent.match(/Android|webOS|iPhone|iPad|iPod|BlackBerry|Windows Phone/i) !== null
    )
}
