// Import only types to avoid adding `@sentry/browser` to our bundle.
import type { Hub, init, onLoad } from '@sentry/browser'

import { authenticatedUser } from '../../auth'
import { shouldErrorBeReported } from '../shouldErrorBeReported'

export type SentrySDK = Hub & {
    init: typeof init
    onLoad: typeof onLoad
}

declare global {
    const Sentry: SentrySDK
}

export function initSentry(): void {
    if (
        typeof Sentry !== 'undefined' &&
        window.context.sentryDSN &&
        (process.env.NODE_ENV === 'production' || process.env.ENABLE_MONITORING)
    ) {
        const { sentryDSN, version } = window.context

        // Wait for Sentry to lazy-load from the script tag defined in the `app.html`.
        // https://sentry-docs-git-patch-1.sentry.dev/platforms/javascript/guides/react/install/lazy-load-sentry/
        Sentry.onLoad(() => {
            Sentry.init({
                dsn: sentryDSN,
                release: 'frontend@' + version,
                beforeSend(event, hint) {
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
        })
    }
}
