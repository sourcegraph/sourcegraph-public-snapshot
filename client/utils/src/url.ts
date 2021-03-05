import { tryCatch } from './errors'

/**
 * Returns true if the given URL points outside the current site.
 */
export const isExternalLink = (url: string): boolean =>
    !!tryCatch(() => new URL(url, window.location.href).origin !== window.location.origin)
