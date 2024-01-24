import { pluralize } from '@sourcegraph/common'

/**
 * Access token scopes, documented at the GraphQL Mutation.createAccessToken.
 */
export enum AccessTokenScopes {
    UserAll = 'user:all',
    SiteAdminSudo = 'site-admin:sudo',
}

/**
 * Converts an array of expiration day values to a record mapping display names
 * to the number of seconds for each value.
 *
 * @param expirationOptions Array of expiration day values
 * @returns Record mapping display name strings to number of seconds
 */
export function getExpirationOptions(expirationOptions: number[]): Record<string, number> {
    const opts: Record<string, number> = {}
    for (const option of expirationOptions) {
        opts[`${option} ${pluralize('day', option)}`] = option * 86400 // convert to seconds
    }
    return opts
}
