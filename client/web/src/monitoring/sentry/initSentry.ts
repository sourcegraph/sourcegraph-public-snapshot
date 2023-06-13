// Import only types to avoid adding `@sentry/browser` to our bundle.
import type { Hub, init, onLoad } from '@sentry/browser'

import { logger } from '@sourcegraph/common'

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
        (process.env.NODE_ENV === 'production' || process.env.ENABLE_SENTRY)
    ) {
        const { sentryDSN, version } = window.context

        try {
            const initSentry = (): void => {
                Sentry.init({
                    dsn: sentryDSN,
                    // TODO frontend platform team, follow-up to https://github.com/sourcegraph/sourcegraph/pull/38411
                    // tunnel: '/-/debug/sentry_tunnel',
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
            }

            // Wait for Sentry to lazy-load from the script tag defined in the `app.html` if it
            // hasn't already.
            //
            // https://sentry-docs-git-patch-1.sentry.dev/platforms/javascript/guides/react/install/lazy-load-sentry/
            if (typeof Sentry.init === 'function') {
                initSentry()
            } else {
                Sentry.onLoad(initSentry)
            }
        } catch (error) {
            logger.error('Error initializing Sentry', error)
        }
    }
}
