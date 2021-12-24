import classNames from 'classnames'
import { parseISO } from 'date-fns'
import formatDistance from 'date-fns/formatDistance'
import CloudDownloadIcon from 'mdi-react/CloudDownloadIcon'
import React, { useMemo } from 'react'
import { Link } from 'react-router-dom'

import { isErrorLike } from '@sourcegraph/common'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { ErrorAlert } from '../components/alerts'
import { PageTitle } from '../components/PageTitle'

import { fetchSiteUpdateCheck } from './backend'
import styles from './SiteAdminUpdatesPage.module.scss'

interface Props extends TelemetryProps {}

/**
 * A page displaying information about available updates for the server.
 */
export const SiteAdminUpdatesPage: React.FunctionComponent<Props> = ({ telemetryService }) => {
    useMemo(() => {
        telemetryService.logViewEvent('SiteAdminUpdates')
    }, [telemetryService])

    const state = useObservable(useMemo(() => fetchSiteUpdateCheck(), []))
    const autoUpdateCheckingEnabled = window.context.site['update.channel'] === 'release'

    if (state === undefined) {
        return <LoadingSpinner />
    }

    const updateCheck = state.updateCheck

    return (
        <div>
            <PageTitle title="Updates - Admin" />
            <h2>Updates</h2>
            {isErrorLike(state) && <ErrorAlert error={state} />}
            {updateCheck && (updateCheck.pending || updateCheck.checkedAt) && (
                <div>
                    {updateCheck.pending && (
                        <div className={classNames('alert alert-primary', styles.alert)}>
                            <LoadingSpinner className="icon-inline" /> Checking for updates... (reload in a few seconds)
                        </div>
                    )}
                    {!updateCheck.errorMessage &&
                        (updateCheck.updateVersionAvailable ? (
                            <div className={classNames('alert alert-success', styles.alert)}>
                                <CloudDownloadIcon className="icon-inline" /> Update available:{' '}
                                <a href="https://about.sourcegraph.com">{updateCheck.updateVersionAvailable}</a>
                            </div>
                        ) : (
                            <div className={classNames('alert alert-success', styles.alert)}>Up to date.</div>
                        ))}
                    {updateCheck.errorMessage && (
                        <ErrorAlert
                            className={styles.alert}
                            prefix="Error checking for updates"
                            error={updateCheck.errorMessage}
                        />
                    )}
                </div>
            )}

            {!autoUpdateCheckingEnabled && (
                <div className={classNames('alert alert-warning', styles.alert)}>
                    Automatic update checking is disabled.
                </div>
            )}

            <p className="site-admin-updates_page__info">
                <small>
                    <strong>Current product version:</strong> {state.productVersion} ({state.buildVersion})
                </small>
                <br />
                <small>
                    <strong>Last update check:</strong>{' '}
                    {updateCheck.checkedAt
                        ? formatDistance(parseISO(updateCheck.checkedAt), new Date(), {
                              addSuffix: true,
                          })
                        : 'never'}
                    .
                </small>
                <br />
                <small>
                    <strong>Automatic update checking:</strong> {autoUpdateCheckingEnabled ? 'on' : 'off'}.{' '}
                    <Link to="/site-admin/configuration">Configure</Link> <code>update.channel</code> to{' '}
                    {autoUpdateCheckingEnabled ? 'disable' : 'enable'}.
                </small>
            </p>
            <p>
                <a href="https://about.sourcegraph.com/changelog" target="_blank" rel="noopener">
                    Sourcegraph changelog
                </a>
            </p>
        </div>
    )
}
