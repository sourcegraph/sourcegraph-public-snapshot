// Import only types to avoid including @sentry/browser into the main chunk.
import type { Hub, init, onLoad } from '@sentry/browser'

import { authenticatedUser } from '../auth'

import { shouldErrorBeReported } from './shouldErrorBeReported'

window.addEventListener('error', error => {
    /**
     * The "ResizeObserver loop limit exceeded" error means that `ResizeObserver` was not
     * able to deliver all observations within a single animation frame. It doesn't break
     * the functionality of the application. The W3C considers converting this error to a warning:
     * https://github.com/w3c/csswg-drafts/issues/5023
     * We can safely ignore it in the production environment to avoid hammering Sentry and other
     * libraries relying on `window.addEventListener('error', callback)`.
     */
    const isResizeObserverLoopError = error.message === 'ResizeObserver loop limit exceeded'

    if (process.env.NODE_ENV === 'production' && isResizeObserverLoopError) {
        error.stopImmediatePropagation()
    }
})

export type SentrySDK = Hub & {
    init: typeof init
    onLoad: typeof onLoad
}

declare global {
    const Sentry: SentrySDK
}

if (typeof Sentry !== 'undefined') {
    // Wait for Sentry to lazy-load from the script tag defined in the app.html.
    // https://sentry-docs-git-patch-1.sentry.dev/platforms/javascript/guides/react/install/lazy-load-sentry/
    Sentry.onLoad(() => {
        // This check is required to please the Typescript compiler 🙂.
        if (window.context.sentryDSN) {
            Sentry.init({
                dsn: window.context.sentryDSN,
                release: 'frontend@' + window.context.version,
                beforeSend(event, hint) {
                    // Report errors only in production environment.
                    if (process.env.NODE_ENV !== 'production') {
                        return null
                    }

                    // Use `originalException` to check if we want to ignore the error.
                    if (!hint || shouldErrorBeReported(hint.originalException)) {
                        return event
                    }

                    return null
                },
            })

            // Sentry is never un-initialized.
            // eslint-disable-next-line rxjs/no-ignored-subscription
            authenticatedUser.subscribe(user => {
                Sentry.configureScope(scope => {
                    if (user) {
                        scope.setUser({ id: user.id })
                    }
                })
            })
        }
    })
}
