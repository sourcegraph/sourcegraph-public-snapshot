// Import only types to avoid adding `@datadog/browser-rum-slim` to our bundle.
import type { RumGlobal } from '@datadog/browser-rum-slim'

import { authenticatedUser } from '../../auth'
import { shouldErrorBeReported } from '../shouldErrorBeReported'

import { isDatadogRumAvailable } from './datadogClient'

declare global {
    const DD_RUM: RumGlobal
}

/**
 * Datadog is initialized only if:
 * 1. The SDK script is included into the `index.html` template (app.html).
 * 2. Datadog RUM is configured using Sourcegraph site configuration.
 * 3. `ENABLE_MONITORING || NODE_ENV === 'production'` to prevent log spam in the development environment.
 */
export function initDatadog(): void {
    if (
        isDatadogRumAvailable() &&
        window.context.datadog &&
        (process.env.NODE_ENV === 'production' || process.env.ENABLE_MONITORING)
    ) {
        const {
            datadog: { applicationId, clientToken },
            version,
        } = window.context

        // The SDK is loaded asynchronously via an async script defined in the `app.html`.
        // https://docs.datadoghq.com/real_user_monitoring/browser/#cdn-async
        DD_RUM.onReady(() => {
            // Initialization parameters: https://docs.datadoghq.com/real_user_monitoring/browser/#configuration
            DD_RUM.init({
                clientToken,
                applicationId,
                env: process.env.NODE_ENV,
                // Sanitize the development version to meet Datadog tagging requirements.
                // https://docs.datadoghq.com/getting_started/tagging/#defining-tags
                version: version.replace('+', '_'),
                // A relative sampling (in percent) to the number of sessions collected.
                // https://docs.datadoghq.com/real_user_monitoring/browser/modifying_data_and_context/?tab=npm#sampling
                sampleRate: 100,
                // We can enable it later after verifying that basic RUM functionality works.
                // https://docs.datadoghq.com/real_user_monitoring/browser/tracking_user_actions
                trackInteractions: false,
                // It's identical to Sentry `beforeSend` hook for now. When we decide to drop
                // one of the services, we can start using more Datadog-specific properties to filter out logs.
                // https://docs.datadoghq.com/real_user_monitoring/browser/modifying_data_and_context/?tab=npm#enrich-and-control-rum-data
                beforeSend(event) {
                    const { type, context } = event

                    // Use `originalException` to check if we want to ignore the error.
                    if (type === 'error') {
                        return shouldErrorBeReported(context?.originalException)
                    }

                    return true
                },
            })

            // Datadog RUM is never un-initialized so there's no need to handle this subscription.
            // eslint-disable-next-line rxjs/no-ignored-subscription
            authenticatedUser.subscribe(user => {
                // Add user information to a RUM session.
                // https://docs.datadoghq.com/real_user_monitoring/browser/modifying_data_and_context/?tab=npm#identify-user-sessions
                if (user) {
                    DD_RUM.setUser(user)
                } else {
                    DD_RUM.removeUser()
                }
            })
        })
    }
}
