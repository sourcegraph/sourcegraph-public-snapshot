import { gql } from '@sourcegraph/http-client'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import * as GQL from '@sourcegraph/shared/src/schema'

import { UserEvent, EventSource } from '../../graphql-operations'

/**
 * Log a user action on the associated self-hosted Sourcegraph instance (allows site admins on a private
 * Sourcegraph instance to see a count of unique users on a daily, weekly, and monthly basis).
 *
 * @deprecated Use logEvent
 */
export const logUserEvent = (
    event: UserEvent,
    uid: string,
    url: string,
    requestGraphQL: PlatformContext['requestGraphQL']
): void => {
    requestGraphQL<GQL.IMutation>({
        request: gql`
            mutation logUserEvent($event: UserEvent!, $userCookieID: String!) {
                logUserEvent(event: $event, userCookieID: $userCookieID) {
                    alwaysNil
                }
            }
        `,
        variables: { event, userCookieID: uid },
        mightContainPrivateInfo: false,
        // Events are best-effort and non-blocking
        // eslint-disable-next-line rxjs/no-ignored-subscription
    }).subscribe({
        error: () => {
            // Swallow errors. If a Sourcegraph instance isn't upgraded, this request may fail
            // (e.g., if CODEINTELINTEGRATION user events aren't yet supported).
            // However, end users shouldn't experience this failure, as their admin is
            // responsible for updating the instance, and has been (or will be) notified
            // that an upgrade is available via site-admin messaging.
        },
    })
}

/**
 * Log a raw user action on the associated Sourcegraph instance
 */
export const logEvent = (
    event: { name: string; userCookieID: string; url: string; argument?: string | {}; publicArgument?: string | {} },
    requestGraphQL: PlatformContext['requestGraphQL']
): void => {
    requestGraphQL<GQL.IMutation>({
        request: gql`
            mutation logEvent(
                $name: String!
                $userCookieID: String!
                $url: String!
                $source: EventSource!
                $argument: String
                $publicArgument: String
            ) {
                logEvent(
                    event: $name
                    userCookieID: $userCookieID
                    url: $url
                    source: $source
                    argument: $argument
                    publicArgument: $publicArgument
                ) {
                    alwaysNil
                }
            }
        `,
        variables: {
            ...event,
            source: EventSource.CODEHOSTINTEGRATION,
            argument: event.argument && JSON.stringify(event.argument),
            publicArgument: event.publicArgument && JSON.stringify(event.publicArgument),
        },
        mightContainPrivateInfo: false,
        // eslint-disable-next-line rxjs/no-ignored-subscription
    }).subscribe({
        error: error => {
            // Swallow errors. If a Sourcegraph instance isn't upgraded, this request may fail
            // (i.e. the new GraphQL API `logEvent` hasn't been added).
            // However, end users shouldn't experience this failure, as their admin is
            // responsible for updating the instance, and has been (or will be) notified
            // that an upgrade is available via site-admin messaging.
        },
    })
}
