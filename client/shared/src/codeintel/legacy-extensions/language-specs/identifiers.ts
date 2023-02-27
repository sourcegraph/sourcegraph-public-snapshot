/**
 * Create an identifier char pattern that matches alphanum and underscore plus any
 * additional extra characters that are supplied.
 *
 * @param extraChars Extra characters to add to pattern. It is expected that
 * any special regex chars are escaped for use in a character class.
 */
export function createIdentCharPattern(extraChars: string): RegExp {
    return new RegExp(`[A-Za-z0-9_${extraChars}]`)
}

/** Matches alphanum, underscore, bang, and question mark. */
export const rubyIdentCharPattern = createIdentCharPattern('!?')
