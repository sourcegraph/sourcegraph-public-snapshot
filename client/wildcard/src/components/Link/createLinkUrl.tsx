<<<<<<< HEAD
export interface URLComponents {
    pathname?: string
    search?: string
    hash?: string
}

/**
 * Convenience method provided for translation of URLComponents
 * into strings that are accepted by the RouterLink component.
 */
export const createLinkUrl = (location: URLComponents): string => {
    const { pathname = '', search, hash } = location

    const components = [pathname]

    if (search?.length) {
        components.push(search.startsWith('?') ? search : `?${search}`)
    }

    if (hash?.length) {
        components.push(hash.startsWith('#') ? hash : `#${hash}`)
    }

    return components.join('')
}
=======
import { createPath, Location } from 'react-router-dom-v5-compat'

/**
 * Convenience method provided for translating Location objects
 * into strings that are accepted by the RouterLink component.
 */
export const createLinkUrl = (location: Partial<Location<unknown>>): string => createPath(location)
>>>>>>> 57f70fbfef (Revert Revert "WIP Router V6 migration: update Link component (#36285)" (#37267)")
