import React from 'react'

import classNames from 'classnames'
import { parseISO } from 'date-fns'
import differenceInDays from 'date-fns/differenceInDays'

import { renderMarkdown } from '@sourcegraph/common'
import { gql, useQuery } from '@sourcegraph/http-client'
import { useSettings } from '@sourcegraph/shared/src/settings/settings'
import { Link, Markdown } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { GlobalAlertsSiteFlagsResult, GlobalAlertsSiteFlagsVariables } from '../graphql-operations'
import { FreeUsersExceededAlert } from '../site/FreeUsersExceededAlert'
import { LicenseExpirationAlert } from '../site/LicenseExpirationAlert'
import { NeedsRepositoryConfigurationAlert } from '../site/NeedsRepositoryConfigurationAlert'
import { siteFlagFieldsFragment } from '../storm/pages/LayoutPage/LayoutPage.loader'

import { GlobalAlert } from './GlobalAlert'
import { Notices, VerifyEmailNotices } from './Notices'

import styles from './GlobalAlerts.module.scss'

interface Props {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphApp: boolean
}

// NOTE: The name of the query is also added in the refreshSiteFlags() function
// found in client/web/src/site/backend.tsx
const QUERY = gql`
    query GlobalAlertsSiteFlags {
        site {
            ...SiteFlagFields
        }
        codeIntelligenceConfigurationPolicies(forEmbeddings: true) {
            totalCount
        }
        repos: repositories(cloned: true) {
            totalCount
        }
        reposWithEmbeddings: repositories(cloned: true, embeddingEnabled: true) {
            totalCount
        }
    }

    ${siteFlagFieldsFragment}
`
/**
 * Fetches and displays relevant global alerts at the top of the page
 */
export const GlobalAlerts: React.FunctionComponent<Props> = ({ authenticatedUser, isSourcegraphApp }) => {
    const settings = useSettings()
    const { data } = useQuery<GlobalAlertsSiteFlagsResult, GlobalAlertsSiteFlagsVariables>(QUERY, {
        fetchPolicy: 'cache-and-network',
    })
    const siteFlagsValue = data?.site

    const showNoEmbeddingPoliciesAlert =
        window.context?.codyEnabled && data?.codeIntelligenceConfigurationPolicies.totalCount === 0

    const reposCount = data?.repos.totalCount ?? 0
    const reposWithEmbeddingsCount = data?.reposWithEmbeddings.totalCount ?? 0
    const showIncreaseReposWithEmbeddingsAlert =
        window.context?.codyEnabled &&
        !showNoEmbeddingPoliciesAlert &&
        reposCount > 0 &&
        reposWithEmbeddingsCount / reposCount < 0.5

    return (
        <div className={classNames('test-global-alert', styles.globalAlerts)}>
            {siteFlagsValue && (
                <>
                    {siteFlagsValue?.externalServicesCounts.remoteExternalServicesCount === 0 && !isSourcegraphApp && (
                        <NeedsRepositoryConfigurationAlert className={styles.alert} />
                    )}
                    {siteFlagsValue.freeUsersExceeded && (
                        <FreeUsersExceededAlert
                            noLicenseWarningUserCount={siteFlagsValue.productSubscription.noLicenseWarningUserCount}
                            className={styles.alert}
                        />
                    )}
                    {siteFlagsValue.alerts.map((alert, index) => (
                        <GlobalAlert key={index} alert={alert} className={styles.alert} />
                    ))}
                    {siteFlagsValue.productSubscription.license &&
                        (() => {
                            const expiresAt = parseISO(siteFlagsValue.productSubscription.license.expiresAt)
                            return (
                                differenceInDays(expiresAt, Date.now()) <= 7 && (
                                    <LicenseExpirationAlert
                                        expiresAt={expiresAt}
                                        daysLeft={Math.floor(differenceInDays(expiresAt, Date.now()))}
                                        className={styles.alert}
                                    />
                                )
                            )
                        })()}
                </>
            )}
            {settings?.motd &&
                Array.isArray(settings.motd) &&
                settings.motd.map(motd => (
                    <DismissibleAlert
                        key={motd}
                        partialStorageKey={`motd.${motd}`}
                        variant="info"
                        className={styles.alert}
                    >
                        <Markdown dangerousInnerHTML={renderMarkdown(motd)} />
                    </DismissibleAlert>
                ))}
            {process.env.SOURCEGRAPH_API_URL && (
                <DismissibleAlert
                    key="dev-web-server-alert"
                    partialStorageKey="dev-web-server-alert"
                    variant="danger"
                    className={styles.alert}
                >
                    <div>
                        <strong>Warning!</strong> This build uses data from the proxied API:{' '}
                        <Link className={styles.proxyLink} target="__blank" to={process.env.SOURCEGRAPH_API_URL}>
                            {process.env.SOURCEGRAPH_API_URL}
                        </Link>
                    </div>
                    .
                </DismissibleAlert>
            )}
            {showNoEmbeddingPoliciesAlert && authenticatedUser?.siteAdmin && (
                <DismissibleAlert
                    key="no-embeddings-policies-alert"
                    partialStorageKey="no-embeddings-policies-alert"
                    variant="danger"
                    className={styles.alert}
                >
                    <div>
                        <strong>Warning!</strong> No embeddings policies have been configured. This will lead to poor
                        results from Cody, Sourcegraph’s AI assistant. Add an{' '}
                        <Link to="/site-admin/embeddings/configuration">embedding policy</Link>
                    </div>
                    .
                </DismissibleAlert>
            )}
            {showIncreaseReposWithEmbeddingsAlert && authenticatedUser?.siteAdmin && (
                <DismissibleAlert
                    key="needs-more-repos-with-embeddings-alert"
                    partialStorageKey="needs-more-repos-with-embeddings-alert"
                    variant="warning"
                    className={styles.alert}
                >
                    <div>
                        <strong>Warning!</strong> Less than 50% of your active repositories are covered by your
                        embedding policy. To provide the best experience to Cody AI assistant users, add more
                        repositories to your{' '}
                        <Link to="/site-admin/embeddings/configuration" className="text-nowrap">
                            embedding policy
                        </Link>
                    </div>
                    .
                </DismissibleAlert>
            )}
            <Notices alertClassName={styles.alert} location="top" />

            {/* The link in the notice doesn't work in the Sourcegraph app since it's rendered by Markdown,
            so don't show it there for now. */}
            {!isSourcegraphApp && (
                <VerifyEmailNotices authenticatedUser={authenticatedUser} alertClassName={styles.alert} />
            )}
        </div>
    )
}
