import { createPath, Location } from 'react-router-dom-v5-compat'

/**
 * Convenience method provided for translating Location objects
 * into strings that are accepted by the RouterLink component.
 */
export const createLinkUrl = (location: Partial<Location<unknown>>): string => createPath(location)
