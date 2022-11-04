/**
 * Feature flag for showing Sourcegraph.com subscriptions and licensing features.
 */
export const SHOW_BUSINESS_FEATURES = Boolean(window.context?.sourcegraphDotComMode)
