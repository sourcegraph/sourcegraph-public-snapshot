import React, { FunctionComponent, useMemo, useState } from 'react'

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

interface Props extends TelemetryProps {
    isSourcegraphApp: boolean
}
const capitalize = (text: string): string => (text && text[0].toUpperCase() + text.slice(1)) || ''

const SiteUpdateCheck: React.FC = () => {
    const { data, loading, error } = useQuery<SiteUpdateCheckResult, SiteUpdateCheckVariables>(SITE_UPDATE_CHECK, {})
    const autoUpdateCheckingEnabled = window.context.site['update.channel'] === 'release'

    return (
        <>
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

            <small>
                {autoUpdateCheckingEnabled
                    ? 'Automatically checking for updates.'
                    : 'Automatic checking for updates disabled.'}{' '}
                Change <Code>update.channel</Code> in <Link to="/site-admin/configuration">site configuration</Link> to{' '}
                {autoUpdateCheckingEnabled ? 'disable' : 'enable'} automatic checking.
            </small>
        </>
    )
}

const SiteUpgradeReadiness: FunctionComponent = () => {
    const { data, loading, error, refetch } = useQuery<SiteUpgradeReadinessResult, SiteUpgradeReadinessVariables>(
        SITE_UPGRADE_READINESS,
        {}
    )
    const [isExpanded, setIsExpanded] = useState(true)
    return (
        <>
            {error && !loading && <ErrorAlert error={error} />}
            {loading && !error && <LoadingSpinner />}
            {data && !loading && (
                <>
                    <div className="d-flex flex-row justify-content-between">
                        <H3>Schema drift</H3>
                        <Button onClick={() => refetch()} variant="primary" size="sm">
                            {' '}
                            Refresh{' '}
                        </Button>
                    </div>

                    {data.site.upgradeReadiness.schemaDrift.length > 0 ? (
                        <Collapse isOpen={isExpanded} onOpenChange={setIsExpanded} openByDefault={false}>
                            <div className="mb-4">
                                <Alert className={classNames('mb-0', styles.alert)} variant="danger">
                                    <span>
                                        There are schema drifts detected. Please follow suggestions below. If you need
                                        assistance, please contact{' '}
                                        <Link
                                            to="mailto:support@sourcegraph.com"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                        >
                                            Sourcegraph support
                                        </Link>
                                        {'. '}
                                    </span>
                                </Alert>
                            </div>

                            <CollapseHeader
                                as={Button}
                                variant="secondary"
                                outline={true}
                                className={classNames(
                                    'p-0 m-0 mt-2 mb-2 border-0 w-100 font-weight-normal d-flex justify-content-between align-items-center',
                                    styles.header
                                )}
                            >
                                <H4 className="m-0">View drift output</H4>
                                <Icon
                                    aria-hidden={true}
                                    svgPath={isExpanded ? mdiChevronUp : mdiChevronDown}
                                    className="mr-1"
                                    size="md"
                                />
                            </CollapseHeader>

                            <CollapsePanel>
                                {data.site.upgradeReadiness.schemaDrift.map(summary => (
                                    <div key={summary.name} className={styles.container}>
                                        <div className={styles.tableContainer}>
                                            <div className={styles.table}>
                                                <div className={styles.label}>Problem:</div>
                                                <div>{summary.problem}</div>
                                            </div>
                                            <div className={styles.table}>
                                                <div className={styles.label}>Solution:</div>
                                                <div>{capitalize(summary.solution)}</div>
                                            </div>
                                            <div className={styles.table}>
                                                <div className={styles.label}>Hint:</div>
                                                <div>
                                                    {summary.urlHint ? (
                                                        <Link to={summary.urlHint}>
                                                            See Sourcegraph query for potential fix
                                                        </Link>
                                                    ) : (
                                                        'Not Applicable'
                                                    )}
                                                </div>
                                            </div>
                                        </div>
                                        <div className={styles.outputContainer}>
                                            <div className={styles.infoContainer}>
                                                <div className={styles.label}>Current Delta:</div>
                                                <div>
                                                    <LogOutput
                                                        text={summary.diff ? summary.diff : 'None'}
                                                        logDescription="The object diff"
                                                    />
                                                </div>
                                            </div>

                                            <div className={styles.infoContainer}>
                                                <div className={styles.label}>Suggested statements to repair:</div>
                                                <div>
                                                    <LogOutput
                                                        text={
                                                            summary.statements ? summary.statements.join('\n') : 'None'
                                                        }
                                                        logDescription="SQL statements to repair"
                                                    />
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </CollapsePanel>
                        </Collapse>
                    ) : (
                        <Alert className={classNames('mb-0', styles.alert)} variant="success">
                            <Text> There is no schema drift detected.</Text>
                        </Alert>
                    )}
                    <hr className="my-3" />
                    <H3>Required out-of-band migrations</H3>
                    {data.site.upgradeReadiness.requiredOutOfBandMigrations.length > 0 ? (
                        <>
                            <span>
                                <Alert className={classNames('mb-0', styles.alert)} variant="warning">
                                    There are pending out-of-band migrations that need to complete, please go to{' '}
                                    <Link to="/site-admin/migrations?filters=pending">migrations</Link> to check
                                    details.
                                </Alert>
                            </span>
                            <ul className="mt-2 pl-3">
                                {data.site.upgradeReadiness.requiredOutOfBandMigrations.map(oobm => (
                                    <li key={oobm.id}>{oobm.description}</li>
                                ))}
                            </ul>
                        </>
                    ) : (
                        <Alert className={classNames('mb-0', styles.alert)} variant="success">
                            <Text className="mb-0">
                                {' '}
                                There are no pending out-of-band migrations that need to complete.
                            </Text>
                        </Alert>
                    )}
                </>
            )}
        </>
    )
}

/**
 * A page displaying information about available updates for the Sourcegraph instance. As well as the readiness status of the instance for upgrade.
 */
export const SiteAdminUpdatesPage: React.FC<Props> = ({ telemetryService, isSourcegraphApp }) => {
    useMemo(() => {
        telemetryService.logViewEvent('SiteAdminUpdates')
    }, [telemetryService])

    return (
        <div>
            <PageTitle title="Updates - Admin" />

            <PageHeader path={[{ text: 'Updates' }]} headingElement="h2" className="mb-3" />
            <Container className="mb-3">
                {isSourcegraphApp ? (
                    <Text className="mb-1">
                        We're making regular improvements to the Sourcegraph app.
                        <br /> For information on how to upgrade to the latest version, see{' '}
                        <Link to="/help/app#upgrading" target="_blank" rel="noopener">
                            our docs
                        </Link>
                        .
                    </Text>
                ) : (
                    <SiteUpdateCheck />
                )}
            </Container>

            {!isSourcegraphApp && (
                <>
                    <PageHeader path={[{ text: 'Upgrade Readiness' }]} headingElement="h2" className="mb-3" />
                    <Container className="mb-3">
                        <SiteUpgradeReadiness />
                    </Container>
                </>
            )}
        </div>
    )
}
