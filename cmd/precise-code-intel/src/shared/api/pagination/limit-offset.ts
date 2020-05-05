/**
 * Normalize limit and offset values extracted from the query string.
 *
 * @param query Parameter bag.
 * @param defaultLimit The limit to use if one is not supplied.
 */
export const extractLimitOffset = (
    {
        limit,
        offset,
    }: {
        /** The limit value extracted from the query string. */
        limit?: number
        /** The offset value extracted from the query string. */
        offset?: number
    },
    defaultLimit: number
): { limit: number; offset: number } => ({ limit: limit || defaultLimit, offset: offset || 0 })
