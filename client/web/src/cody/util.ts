// The URL to direct users in order to manage their Cody Pro subscription.
export const manageSubscriptionRedirectURL = 'https://accounts.sourcegraph.com/cody/subscription'

/**
 * useEmbeddedCodyProUi returns if we expect the Cody Pro UI to be served from sourcegraph.com. Meaning
 * we should direct the user to `/cody/manage/subscription` for making changes.
 *
 * If false, we rely on the current behavior. Where users are directed to https://accounts.sourcegraph.com/cody
 * for managing their Cody Pro subscription information.
 */
export function useEmbeddedCodyProUi(): boolean {
    const codyProConfig = window.context.frontendCodyProConfig
    if (codyProConfig && codyProConfig.stripePublishableKey) {
        return true
    }
    return false
}
