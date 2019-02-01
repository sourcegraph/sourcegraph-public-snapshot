import { numberWithCommas, pluralize } from '../util/strings'

/**
 * Returns "N users" (properly pluralized and with commas added to N as needed).
 */
export function formatUserCount(userCount: number, hyphenate?: boolean): string {
    if (hyphenate) {
        return `${numberWithCommas(userCount)}-user`
    }
    return `${numberWithCommas(userCount)} ${pluralize('user', userCount)}`
}
