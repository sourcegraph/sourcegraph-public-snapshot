import { CodyProRoutes } from './codyProRoutes'

// URL the user needs to navigate to in order to modify their Cody Pro subscription.
export const manageSubscriptionRedirectURL = `${
    window.context?.frontendCodyProConfig?.sscBaseUrl || 'https://accounts.sourcegraph.com/cody'
}/subscription`

/**
 * useEmbeddedCodyProUi returns if we expect the Cody Pro UI to be served from sourcegraph.com. Meaning
 * we should direct the user to `/cody/manage/subscription` for making changes.
 *
 * If false, we rely on the current behavior. Where users are directed to https://accounts.sourcegraph.com/cody
 * for managing their Cody Pro subscription information.
 */
export function isEmbeddedCodyProUIEnabled(): boolean {
    return !!window.context.frontendCodyProConfig?.useEmbeddedUI
}

/**
 * getManageSubscriptionPageURL returns the URL to direct the user to in order to manage their Cody Pro subscription.
 */
export function getManageSubscriptionPageURL(): string {
    return isEmbeddedCodyProUIEnabled() ? CodyProRoutes.SubscriptionManage : manageSubscriptionRedirectURL
}

/**
 * Note that this is a very simplistic approach.
 * "doesThisStringRoughlyResembleAnEmailAddress" would be a more accurate name.
 * And it is definitely not meant to replace the backend validation.
 */
export function isValidEmailAddress(emailAddress: string): boolean {
    return emailRegex.test(emailAddress)
}

/**
 * Regular expression to validate whether a string looks like an email address:
 *  - Contains a single "@" that is not at the beginning or at the end.
 *  - Contains at least one "." after the "@" that is not at the end.
 *
 * NOTE: Keep this in sync with `emailRegex` in the backend
 * (https://sourcegraph.sourcegraph.com/search?q=context:global+r:github.com/sourcegraph/sourcegraph-accounts+f:backend/internal/graph/*+%22var+emailRegex+%3D+regexp.%22&patternType=newStandardRC1&sm=1),
 * and keep in mind that the backend validation has the final say, validation in the web app is only for UX improvement.
 */
const emailRegex = /^[^@]+@[^@]+\.[^@]+$/

/**
 * Whether the current user is unable to use Cody because they must verify their email address.
 */
export function currentUserRequiresEmailVerificationForCody(): boolean {
    return window.context?.codyRequiresVerifiedEmail && !window.context?.currentUser?.hasVerifiedEmail
}
