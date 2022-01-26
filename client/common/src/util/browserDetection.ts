export function isFirefox(): boolean {
    return window.navigator.userAgent.includes('Firefox')
}

export const isMacPlatform = window.navigator.platform.includes('Mac')
