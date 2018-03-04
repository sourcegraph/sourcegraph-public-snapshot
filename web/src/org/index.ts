import { VALID_USERNAME_REGEXP } from '../user'

/**
 * Regular expression to identify valid organization names.
 */
export const VALID_ORG_NAME_REGEXP = VALID_USERNAME_REGEXP

/** Returns the URL path to an organization's public profile */
export function orgURL(name: string): string {
    return `/organizations/${name}`
}
