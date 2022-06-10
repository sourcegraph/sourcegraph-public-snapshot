import { createPath, LocationDescriptorObject } from 'history'

/**
 * Convenience method provided for translation of History.LocationDescriptorObject's
 * into strings that are accepted by the RouterLink component.
 */
export const createLinkUrl = (location: LocationDescriptorObject): string => createPath(location)
