import React, { useMemo } from 'react'

import { mdiOpenInNew, mdiCheckCircle } from '@mdi/js'
import classNames from 'classnames'
import { parseISO } from 'date-fns'
import formatDistance from 'date-fns/formatDistance'
import { SiteUpdateCheckResult, SiteUpdateCheckVariables } from 'src/graphql-operations'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, Link, PageHeader, Alert, Icon, Code, Container, Text } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'

import { SITE_UPDATE_CHECK } from './backend'

import styles from './SiteAdminUpdatesPage.module.scss'

interface Props extends TelemetryProps {}

/**
 * A page displaying information about available updates for the server.
 */
export const SiteAdminUpdatesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ telemetryService }) => {
    useMemo(() => {
        telemetryService.logViewEvent('SiteAdminUpdates')
    }, [telemetryService])

    const { data, loading, error } = useQuery<SiteUpdateCheckResult, SiteUpdateCheckVariables>(SITE_UPDATE_CHECK, {})
    const autoUpdateCheckingEnabled = window.context.site['update.channel'] === 'release'

    return (
        <div>
            <PageTitle title="Updates - Admin" />
            <PageHeader path={[{ text: 'Updates' }]} headingElement="h2" className="mb-3" />

            <Container>
                {error && !loading && <ErrorAlert error={error} />}
                {loading && !error && <LoadingSpinner />}
                {data && (
                    <>
                        <Text className="mb-2">
                            Version {data.site.productVersion}{' '}
                            <small className="text-muted">
                                (
                                <Link to="https://about.sourcegraph.com/changelog" target="_blank" rel="noopener">
                                    changelog
                                </Link>
                                )
                            </small>
                            <br />
                        </Text>

                        <div>
                            {data.site.updateCheck.pending && (
                                <Alert className={styles.alert} variant="primary">
                                    <LoadingSpinner /> Checking for updates... (reload in a few seconds)
                                </Alert>
                            )}
                            {data.site.updateCheck.errorMessage && (
                                <ErrorAlert
                                    className={styles.alert}
                                    prefix="Error checking for updates"
                                    error={data.site.updateCheck.errorMessage}
                                />
                            )}
                            {!data.site.updateCheck.errorMessage && (
                                <small>
                                    {data.site.updateCheck.updateVersionAvailable ? (
                                        <Link to="https://about.sourcegraph.com">
                                            Update available to version {data.site.updateCheck.updateVersionAvailable}{' '}
                                            <Icon aria-hidden={true} svgPath={mdiOpenInNew} />
                                        </Link>
                                    ) : (
                                        <span>
                                            <Icon
                                                aria-hidden={true}
                                                className="text-success mr-1"
                                                svgPath={mdiCheckCircle}
                                            />
                                            Up to date
                                        </span>
                                    )}
                                    <span className={classNames('text-muted pl-2 ml-2', styles.lastChecked)}>
                                        {data.site.updateCheck.checkedAt
                                            ? `Last checked ${formatDistance(
                                                  parseISO(data.site.updateCheck.checkedAt),
                                                  new Date(),
                                                  {
                                                      addSuffix: true,
                                                  }
                                              )}`
                                            : 'Never checked for updates'}
                                    </span>
                                </small>
                            )}
                        </div>
                    </>
                )}
            </Container>

            <small>
                {autoUpdateCheckingEnabled
                    ? 'Automatically checking for updates.'
                    : 'Automatic checking for updates disabled.'}{' '}
                Change <Code>update.channel</Code> in <Link to="/site-admin/configuration">site configuration</Link> to{' '}
                {autoUpdateCheckingEnabled ? 'disable' : 'enable'} automatic checking.
            </small>
        </div>
    )
}
