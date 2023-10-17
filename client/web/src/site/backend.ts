import type { ApolloClient } from '@apollo/client'

/**
 * Utility helper to refetch all site flag related queries.
 */
export async function refreshSiteFlags(client: ApolloClient<{}>): Promise<void> {
    await client.refetchQueries({
        include: ['GlobalAlertsSiteFlags', 'UserSettingsEmailsSiteFlags', 'SiteConfig'],
    })
}
