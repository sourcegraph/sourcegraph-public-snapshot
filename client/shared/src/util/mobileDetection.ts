// Function transcribed from https://developer.mozilla.org/en-US/docs/Web/HTTP/Browser_detection_using_the_user_agent (Mobile device detection)
export function hasTouchScreen(): boolean {
    if ('maxTouchPoints' in navigator) {
        return navigator.maxTouchPoints > 0
    }
    const matchMediaQuery = 'matchMedia' in window && window.matchMedia('(pointer:coarse)')
    if (matchMediaQuery && matchMediaQuery.media === '(pointer:coarse)') {
        return matchMediaQuery.matches
    }
    if ('orientation' in window) {
        return true // deprecated, but good fallback
    }
    // Only as a last resort, fall back to user agent sniffing
    return /\b(blackberry|webos|iphone|iemobile|android|windows phone|ipad|ipod)\b/i.test(navigator.userAgent)
}
