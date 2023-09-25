import { gql } from '@sourcegraph/http-client'
import type { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import {
    TelemetryEventSourceInput,
    TelemetryEventParametersInput,
    TelemetryEventMarketingTrackingInput,
    type recordEventResult,
} from '../../graphql-operations'

/**
 * Record a raw user action on the associated Sourcegraph instance
 */
export const recordEvent = (
    event: {
        action: string
        feature: string
        source: TelemetryEventSourceInput | {}
        parameters: TelemetryEventParametersInput | {}
        marketingTracking?: TelemetryEventMarketingTrackingInput | {}
    },
    requestGraphQL: PlatformContext['requestGraphQL']
): void => {
    requestGraphQL<recordEventResult>({
        request: gql`
            mutation recordEvent(
                $feature: String!
                $action: String!
                $source: TelemetryEventSourceInput!
                $parameters: TelemetryEventParametersInput!
                $marketingTracking: TelemetryEventMarketingTrackingInput
            ) {
                recordEvent(
                    feature: $feature
                    action: $action
                    source: $source
                    parameters: $parameters
                    marketingTracking: $marketingTracking
                ) {
                    alwaysNil
                }
            }
        `,
        variables: {
            ...event,
        },
        mightContainPrivateInfo: false,
        // eslint-disable-next-line rxjs/no-ignored-subscription
    }).subscribe({
        error: error => {
            // Swallow errors. If a Sourcegraph instance isn't upgraded, this request may fail
            // (i.e. the new GraphQL API `recordEvent` hasn't been added).
            // However, end users shouldn't experience this failure, as their admin is
            // responsible for updating the instance, and has been (or will be) notified
            // that an upgrade is available via site-admin messaging.
        },
    })
}
