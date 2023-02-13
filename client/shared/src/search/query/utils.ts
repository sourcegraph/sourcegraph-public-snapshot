import type { Token } from './token'

/**
 * Returns the (string) length of the provided token.
 */
export function getTokenLength(token: Token): number {
    return token.range.end - token.range.start
}
