/**
 * The publishable (non-secret) API key for the billing system.
 */
export const billingPublishableKey: string | undefined = (window as any).context.billingPublishableKey

/**
 * Feature flag for showing Sourcegraph.com subscriptions, licensing, and billing features.
 */
export const USE_DOTCOM_BUSINESS: boolean = Boolean(
    localStorage.getItem('business') !== null ||
        ((window as any).context.sourcegraphDotComMode && billingPublishableKey)
)
