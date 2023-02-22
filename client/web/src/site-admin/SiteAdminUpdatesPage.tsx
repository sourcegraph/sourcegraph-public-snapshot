import React, { FunctionComponent, useMemo, useState } from 'react'

import { ApolloError } from '@apollo/client'
import { mdiOpenInNew, mdiCheckCircle, mdiChevronUp, mdiChevronDown } from '@mdi/js'
import classNames from 'classnames'
import { parseISO } from 'date-fns'
import formatDistance from 'date-fns/formatDistance'
import {
    SiteUpdateCheckResult,
    SiteUpdateCheckVariables,
    SiteUpgradeReadinessResult,
    SiteUpgradeReadinessVariables,
} from 'src/graphql-operations'

import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    LoadingSpinner,
    Link,
    PageHeader,
    Alert,
    Icon,
    Code,
    Container,
    Text,
    ErrorAlert,
    Button,
    Collapse,
    CollapseHeader,
    CollapsePanel,
    H3,
    H4,
} from '@sourcegraph/wildcard'

import { LogOutput } from '../components/LogOutput'
import { PageTitle } from '../components/PageTitle'

import { SITE_UPDATE_CHECK, SITE_UPGRADE_READINESS } from './backend'

import styles from './SiteAdminUpdatesPage.module.scss'

interface Props extends TelemetryProps {}

interface SiteUpdateCheckProps {
    data: SiteUpdateCheckResult | undefined
    loading: boolean
    error: ApolloError | undefined
    autoUpdateCheckingEnabled: boolean
}

const SiteUpdateCheck: FunctionComponent<React.PropsWithChildren<SiteUpdateCheckProps>> = ({
    data,
    loading,
    error,
    autoUpdateCheckingEnabled,
}) => (
    <>
        <PageHeader path={[{ text: 'Updates' }]} headingElement="h2" className="mb-3" />

        <Container className="mb-2">
            {error && !loading && <ErrorAlert error={error} />}
            {loading && !error && <LoadingSpinner />}
            {data && (
                <>
                    <Text className="mb-1">
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
    </>
)

interface SiteUpgradeReadinessProps {
    data: SiteUpgradeReadinessResult | undefined
    loading: boolean
    error: ApolloError | undefined
}

const SiteUpgradeReadiness: FunctionComponent<React.PropsWithChildren<SiteUpgradeReadinessProps>> = ({
    data,
    loading,
    error,
}) => {
    const [isExpanded, setIsExpanded] = useState(true)
    return (
        <>
            <PageHeader path={[{ text: 'Upgrade readiness' }]} headingElement="h2" className="mb-3 mt-3" />
            <Container className="mb-3">
                {error && !loading && <ErrorAlert error={error} />}
                {loading && !error && <LoadingSpinner />}
                {data && (
                    <>
                        <H3 as={H4}>Schema drift</H3>
                        {data.site.upgradeReadiness.schemaDrift.length > 0 ? (
                            <Collapse isOpen={isExpanded} onOpenChange={setIsExpanded} openByDefault={false}>
                                <CollapseHeader
                                    as={Button}
                                    className="p-0 m-0 mb-2 border-0 w-100 font-weight-normal d-flex justify-content-between align-items-center"
                                >
                                    <span>❌ There are schema drifts detected.</span>
                                    <Icon
                                        aria-hidden={true}
                                        svgPath={isExpanded ? mdiChevronUp : mdiChevronDown}
                                        className="mr-1"
                                    />
                                </CollapseHeader>
                                <CollapsePanel>
                                    <LogOutput
                                        text={data.site.upgradeReadiness.schemaDrift}
                                        logDescription="Drift details:"
                                    />
                                </CollapsePanel>
                            </Collapse>
                        ) : (
                            <Text>✅ There is no schema drift detected.</Text>
                        )}

                        <H3 as={H4} className="mt-3">
                            Required out-of-band migrations
                        </H3>
                        {data.site.upgradeReadiness.requiredOutOfBandMigrations.length > 0 ? (
                            <>
                                <span>
                                    ❌ There are pending out-of-band migrations that are deprecated in the current
                                    version, please go to{' '}
                                    <Link to="/site-admin/migrations?filters=pending">migrations</Link> to check
                                    details.
                                </span>
                                <ul className="mt-2 pl-3">
                                    {data.site.upgradeReadiness.requiredOutOfBandMigrations.map(oobm => (
                                        <li key={oobm.id}>{oobm.description}</li>
                                    ))}
                                </ul>
                            </>
                        ) : (
                            <Text>
                                ✅ There are no pending out-of-band migrations that are deprecated in the current
                                version.
                            </Text>
                        )}
                    </>
                )}
            </Container>
        </>
    )
}

/**
 * A page displaying information about available updates for the server.
 */
export const SiteAdminUpdatesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ telemetryService }) => {
    useMemo(() => {
        telemetryService.logViewEvent('SiteAdminUpdates')
    }, [telemetryService])

    const {
        data: updateData,
        loading: updateLoading,
        error: updateError,
    } = useQuery<SiteUpdateCheckResult, SiteUpdateCheckVariables>(SITE_UPDATE_CHECK, {})
    const autoUpdateCheckingEnabled = window.context.site['update.channel'] === 'release'

    const {
        data: upgradeData,
        loading: upgradeLoading,
        error: upgradeError,
    } = useQuery<SiteUpgradeReadinessResult, SiteUpgradeReadinessVariables>(SITE_UPGRADE_READINESS, {})

    return (
        <div>
            <PageTitle title="Updates - Admin" />
            <SiteUpdateCheck
                data={updateData}
                loading={updateLoading}
                error={updateError}
                autoUpdateCheckingEnabled={autoUpdateCheckingEnabled}
            />
            <SiteUpgradeReadiness data={upgradeData} loading={upgradeLoading} error={upgradeError} />
        </div>
    )
}
