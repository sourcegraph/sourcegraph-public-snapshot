export function isFirefox(): boolean {
    return window.navigator.userAgent.includes('Firefox')
}

/**
 * Change isMacPlatform to a function so window
 * is accessed only when the function is called
 */
export function isMacPlatform(): boolean {
    return window.navigator.platform.includes('Mac')
}
