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
    // Examples: 'MacIntel', 'MacPPC', 'Mac68K'
    return typeof window !== 'undefined' && window.navigator.platform.includes('Mac')
}

export function isLinuxPlatform(): boolean {
    // Examples: 'Linux', 'Linux x86_64', 'Linux i686'
    return typeof window !== 'undefined' && window.navigator.platform.includes('Linux')
}

export function isWindowsPlatform(): boolean {
    // Examples: 'Win32', 'Windows', 'Win64'
    return typeof window !== 'undefined' && window.navigator.platform.includes('Win')
}

export function getPlatform(): 'windows' | 'mac' | 'linux' | 'other' {
    return isWindowsPlatform() ? 'windows' : isMacPlatform() ? 'mac' : isLinuxPlatform() ? 'linux' : 'other'
}
