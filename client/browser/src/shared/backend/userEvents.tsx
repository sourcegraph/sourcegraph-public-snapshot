import { DEFAULT_SOURCEGRAPH_URL, repoUrlCache, sourcegraphUrl } from '../util/context'
import { getContext } from './context'
import { mutateGraphQL } from './graphql'

/**
 * Log a user action on the associated self-hosted Sourcegraph instance (allows site admins on a private
 * Sourcegraph instance to see a count of unique users on a daily, weekly, and monthly basis).
 *
 * This is never sent to Sourcegraph.com (i.e., when using the integration with open source code).
 */
export const logUserEvent = (event: string, uid: string): void => {
    const ctx = getContext({ isRepoSpecific: true })
    const url = repoUrlCache[ctx.repoKey] || sourcegraphUrl
    // Only send the request if this is a private, self-hosted Sourcegraph instance.
    if (!url || url === DEFAULT_SOURCEGRAPH_URL) {
        return
    }
    mutateGraphQL({
        ctx,
        request: `mutation logUserEvent($event: UserEvent!, $userCookieID: String!) {
            logUserEvent(event: $event, userCookieID: $userCookieID) {
                alwaysNil
            }
        }`,
        variables: { event, userCookieID: uid },
        retry: false,
        // tslint:disable-next-line: rxjs-no-ignored-subscription
    }).subscribe({
        error: error => {
            // Swallow errors. If a Sourcegraph instance isn't upgraded, this request may fail
            // (e.g., if CODEINTELINTEGRATION user events aren't yet supported).
            // However, end users shouldn't experience this failure, as their admin is
            // responsible for updating the instance, and has been (or will be) notified
            // that an upgrade is available via site-admin messaging.
        },
    })
}
