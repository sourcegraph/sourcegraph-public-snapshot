/**
 * Regular expression to identify valid username.
 */
export const VALID_USERNAME_REGEXP = /^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}$/

/** Returns the URL path to a user's public profile */
export function userURL(username: string): string {
    return `/users/${username}`
}
