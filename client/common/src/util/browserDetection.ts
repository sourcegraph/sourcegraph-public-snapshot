export function isChrome(): boolean {
    return typeof window !== 'undefined' && !!window.navigator.userAgent.match(/chrome|chromium|crios/i)
}

export function isSafari(): boolean {
    return typeof window !== 'undefined' && !!window.navigator.userAgent.match(/safari/i) && !isChrome()
}

export function isFirefox(): boolean {
    return typeof window !== 'undefined' && window.navigator.userAgent.includes('Firefox')
}

export function getBrowserName(): 'chrome' | 'safari' | 'firefox' | 'other' {
    return isChrome() ? 'chrome' : isSafari() ? 'safari' : isFirefox() ? 'firefox' : 'other'
}

/**
 * Change isMacPlatform to a function so window
 * is accessed only when the function is called
 */
export function isMacPlatform(): boolean {
    return typeof window !== 'undefined' && window.navigator.platform.includes('Mac')
}
