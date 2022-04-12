export function isChrome(): boolean {
    return !!window.navigator.userAgent.match(/chrome|chromium|crios/i)
}

export function isSafari(): boolean {
    return !!window.navigator.userAgent.match(/safari/i)
}

export function isFirefox(): boolean {
    return window.navigator.userAgent.includes('Firefox')
}

export function getBrowserName(): 'chrome' | 'safari' | 'firefox' | 'other' {
    return isChrome() ? 'chrome' : isSafari() ? 'safari' : isFirefox() ? 'firefox' : 'other'
}

/**
 * Change isMacPlatform to a function so window
 * is accessed only when the function is called
 */
export function isMacPlatform(): boolean {
    return window.navigator.platform.includes('Mac')
}
