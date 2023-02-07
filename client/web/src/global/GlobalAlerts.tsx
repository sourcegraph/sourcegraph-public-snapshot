import React, { useMemo } from 'react'

import classNames from 'classnames'
import { parseISO } from 'date-fns'
import differenceInDays from 'date-fns/differenceInDays'
import { without } from 'lodash'
import { useLocation } from 'react-router-dom-v5-compat'

import { renderMarkdown } from '@sourcegraph/common'
import { gql, useQuery } from '@sourcegraph/http-client'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { isSettingsValid, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { Link, useObservable, Markdown, Text } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { DismissibleAlert, isAlertDismissed } from '../components/DismissibleAlert'
import { NotConnectedExternalAccountsResult, NotConnectedExternalAccountsVariables } from '../graphql-operations'
import { siteFlags } from '../site/backend'
import { DockerForMacAlert } from '../site/DockerForMacAlert'
import { FreeUsersExceededAlert } from '../site/FreeUsersExceededAlert'
import { LicenseExpirationAlert } from '../site/LicenseExpirationAlert'
import { NeedsRepositoryConfigurationAlert } from '../site/NeedsRepositoryConfigurationAlert'

import { GlobalAlert } from './GlobalAlert'
import { Notices, VerifyEmailNotices } from './Notices'

import styles from './GlobalAlerts.module.scss'

interface Props extends SettingsCascadeProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

/**
 * Fetches and displays relevant global alerts at the top of the page
 */
export const GlobalAlerts: React.FunctionComponent<Props> = ({
    authenticatedUser,
    settingsCascade,
    isSourcegraphDotCom,
}) => {
    const siteFlagsValue = useObservable(siteFlags)

    const verifyEmailProps = useMemo(() => {
        if (!authenticatedUser || !isSourcegraphDotCom) {
            return
        }
        return {
            emails: authenticatedUser.emails.filter(({ verified }) => !verified).map(({ email }) => email),
            settingsURL: authenticatedUser.settingsURL as string,
        }
    }, [authenticatedUser, isSourcegraphDotCom])
    return (
        <div className={classNames('test-global-alert', styles.globalAlerts)}>
            {siteFlagsValue && (
                <>
                    {siteFlagsValue.needsRepositoryConfiguration && (
                        <NeedsRepositoryConfigurationAlert className={styles.alert} />
                    )}
                    {siteFlagsValue.freeUsersExceeded && (
                        <FreeUsersExceededAlert
                            noLicenseWarningUserCount={siteFlagsValue.productSubscription.noLicenseWarningUserCount}
                            className={styles.alert}
                        />
                    )}
                    {/* Only show if the user has already added repositories; if not yet, the user wouldn't experience any Docker for Mac perf issues anyway. */}
                    {window.context.likelyDockerOnMac && window.context.deployType === 'docker-container' && (
                        <DockerForMacAlert className={styles.alert} />
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
            {isSettingsValid<Settings>(settingsCascade) &&
                settingsCascade.final.motd &&
                Array.isArray(settingsCascade.final.motd) &&
                settingsCascade.final.motd.map(motd => (
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
            <Notices alertClassName={styles.alert} location="top" settingsCascade={settingsCascade} />
            {!!verifyEmailProps?.emails.length && (
                <VerifyEmailNotices alertClassName={styles.alert} {...verifyEmailProps} />
            )}
            {!isSourcegraphDotCom && authenticatedUser?.settingsURL && (
                <ConnectExternalAccountsAlert settingsURL={authenticatedUser.settingsURL} />
            )}
        </div>
    )
}

// TODO: move to a separate file
const QUERY = gql`
    query NotConnectedExternalAccounts {
        currentUser {
            externalAccounts {
                nodes {
                    serviceID
                }
            }
        }
        site {
            authProviders {
                nodes {
                    serviceID
                    isBuiltin
                }
            }
        }
    }
`
interface ConnectExternalAccountsAlertProps {
    settingsURL: string
}
const partialStorageKey = 'connect-external-accounts'
const ConnectExternalAccountsAlert: React.FunctionComponent<ConnectExternalAccountsAlertProps> = ({ settingsURL }) => {
    const connectExternalAccountsURL = settingsURL + '/security'
    const location = useLocation()
    const skipQuery = useMemo(
        () => isAlertDismissed(partialStorageKey) || location.pathname === connectExternalAccountsURL,
        [connectExternalAccountsURL, location.pathname]
    )
    const { data } = useQuery<NotConnectedExternalAccountsResult, NotConnectedExternalAccountsVariables>(QUERY, {
        fetchPolicy: 'network-only',
        skip: skipQuery,
    })

    const notConnectedAuthProviders = useMemo(() => {
        const authProviderIDs = data?.site?.authProviders?.nodes
            .filter(({ isBuiltin }) => !isBuiltin)
            .map(({ serviceID }) => serviceID)
        const connectedAuthProviderIDs =
            data?.currentUser?.externalAccounts?.nodes.map(({ serviceID }) => serviceID) ?? []
        return without(authProviderIDs, ...connectedAuthProviderIDs)
    }, [data])

    if (notConnectedAuthProviders.length === 0) {
        return null
    }

    return (
        <DismissibleAlert partialStorageKey={partialStorageKey} variant="info">
            <Text className="mb-0">
                <Link to={connectExternalAccountsURL}>Connect</Link> your account with a third-party login service
                {notConnectedAuthProviders.length > 1 && 's'} to make signing in easier and enable{' '}
                <Link to="/help/admin/repo/permissions#permissions-for-multiple-code-hosts" target="_blank">
                    multiple code host permissions
                </Link>{' '}
                syncing.
            </Text>
        </DismissibleAlert>
    )
}
