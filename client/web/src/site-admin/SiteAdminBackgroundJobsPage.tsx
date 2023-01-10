import React, { useEffect, useMemo, useState } from 'react'

import { mdiAccountHardHat, mdiAlert, mdiCached, mdiDatabase, mdiHelp, mdiNumeric, mdiShape } from '@mdi/js'
import { RouteComponentProps } from 'react-router'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { pluralize } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Button,
    Container,
    ErrorAlert,
    Icon,
    LoadingSpinner,
    PageHeader,
    Text,
    Tooltip,
    useSessionStorage,
} from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'
import { BackgroundJobsResult, BackgroundJobsVariables } from '../graphql-operations'
import { formatDurationLong, formatDurationStructured, StructuredDuration } from '../util/time'

import { ValueLegendItem, ValueLegendList, ValueLegendListProps } from './analytics/components/ValueLegendList'
import { BACKGROUND_JOBS, BACKGROUND_JOBS_PAGE_POLL_INTERVAL_MS } from './backend'

import styles from './SiteAdminBackgroundJobsPage.module.scss'

export interface SiteAdminBackgroundJobsPageProps extends RouteComponentProps, TelemetryProps {}

export type BackgroundJob = BackgroundJobsResult['backgroundJobs']['nodes'][0]

type RunLengthCategory = 'short' | 'long' | 'dangerous'

export const SiteAdminBackgroundJobsPage: React.FunctionComponent<
    React.PropsWithChildren<SiteAdminBackgroundJobsPageProps>
> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminBackgroundJobs')
    }, [telemetryService])

    const { data, loading, error, stopPolling, startPolling } = useQuery<BackgroundJobsResult, BackgroundJobsVariables>(
        BACKGROUND_JOBS,
        {
            variables: { recentRunCount: 5 },
            pollInterval: BACKGROUND_JOBS_PAGE_POLL_INTERVAL_MS,
        }
    )

    return (
        <div>
            <PageTitle title="Background jobs - Admin" />
            <PageHeader
                path={[{ text: 'Background jobs' }]}
                headingElement="h2"
                description="This page lists all running jobs, their routines, recent runs, any errors, timings, and stats."
                className="mb-3"
            />
            <Text>Terminology:</Text>
            <ul>
                <li>
                    <strong>Job</strong>: a bag of routines, started when the Sourcegraph app is launched
                </li>
                <li>
                    <strong>Routine</strong>: a background process that repeatedly executes its task indefinitely, using
                    an interval passed at start
                </li>
                <li>
                    <strong>Run</strong>: a single execution of a routine's task
                </li>
                <li>
                    <strong>Host</strong>: a Sourcegraph instance that starts some jobs when launched
                </li>
                <li>
                    <strong>Instance</strong>: a job ran on a host
                </li>
            </ul>
            <Container className="mb-3">
                {error && !loading && <ErrorAlert error={error} />}
                {loading && !error && <LoadingSpinner />}
                {!loading && !error && data?.backgroundJobs.nodes && (
                    <JobList jobs={data.backgroundJobs.nodes} stopPolling={stopPolling} startPolling={startPolling} />
                )}
            </Container>
        </div>
    )
}

const JobList: React.FunctionComponent<{
    jobs: BackgroundJob[]
    stopPolling: () => void
    startPolling: (pollInterval: number) => void
}> = ({ jobs, stopPolling, startPolling }) => {
    const [polling, setPolling] = useState(true)
    const [onlyShowProblematic, setOnlyShowProblematic] = useSessionStorage(
        'site-admin.background-jobs.only-show-problematic-routines',
        false
    )

    const hostNames = useMemo(
        () =>
            jobs
                .map(job => job.routines[0]?.instances[0]?.hostName)
                .filter((host, index, hosts) => hosts.indexOf(host) === index)
                .filter(host => !!host),
        [jobs]
    )

    const jobsToDisplay = onlyShowProblematic
        ? jobs.filter(job => job.routines.some(routine => isRoutineProblematic(routine)))
        : jobs

    return (
        <>
            <div className="text-right mb-4">
                <Button variant="secondary" onClick={() => setOnlyShowProblematic(!onlyShowProblematic)}>
                    {onlyShowProblematic ? 'Show all routines' : 'Only show problematic routines'}
                </Button>
                <Button
                    variant="secondary"
                    onClick={() => {
                        if (polling) {
                            stopPolling()
                        } else {
                            startPolling(BACKGROUND_JOBS_PAGE_POLL_INTERVAL_MS)
                        }
                        setPolling(!polling)
                    }}
                    className="ml-2"
                >
                    {polling ? 'Pause polling' : 'Resume polling'}
                </Button>
            </div>
            <LegendList jobs={jobsToDisplay} hostNameCount={hostNames.length} />
            {jobsToDisplay ? (
                <ul className="list-group list-group-flush">
                    {jobsToDisplay.map(job => {
                        const jobHostNames = [
                            ...new Set(
                                job.routines.map(routine => routine.instances.map(instance => instance.hostName)).flat()
                            ),
                        ].sort()
                        return (
                            <li key={job.name} className="list-group-item px-0 py-2">
                                <div className="d-flex align-items-center justify-content-between mb-2">
                                    <div className="d-flex flex-row align-items-center mb-0">
                                        <Icon aria-hidden={true} svgPath={mdiAccountHardHat} />{' '}
                                        <Text className="mb-0 ml-2">
                                            <strong>{job.name}</strong> (starts {job.routines.length}{' '}
                                            {pluralize('routine', job.routines.length)}
                                            {hostNames.length > 1
                                                ? ` on ${jobHostNames.length} ${pluralize(
                                                      'instance',
                                                      jobHostNames.length
                                                  )}`
                                                : ''}
                                            )
                                        </Text>
                                    </div>
                                </div>
                                {job.routines
                                    .filter(routine => (onlyShowProblematic ? isRoutineProblematic(routine) : true))
                                    .map(routine => (
                                        <RoutineComponent routine={routine} key={routine.name} />
                                    ))}
                            </li>
                        )
                    })}
                </ul>
            ) : (
                'No jobs to display.'
            )}
        </>
    )
}

const LegendList: React.FunctionComponent<{ jobs: BackgroundJob[]; hostNameCount: number }> = ({
    jobs,
    hostNameCount,
}) => {
    const legends = useMemo<ValueLegendListProps['items']>(() => {
        const routineCount = jobs.reduce((acc, job) => acc + job.routines.length, 0)
        const routineInstanceCount = jobs.reduce(
            (acc, job) => acc + job.routines.reduce((acc, routine) => acc + routine.instances.length, 0),
            0
        )
        const recentRunCount = jobs.reduce(
            (acc, job) => acc + job.routines.reduce((acc, routine) => acc + routine.recentRuns.length, 0),
            0
        )
        const recentRunErrors = jobs.reduce(
            (acc, job) =>
                acc +
                job.routines.reduce((acc, routine) => acc + routine.recentRuns.filter(run => run.error).length, 0),
            0
        )
        const runsForStats = jobs.reduce(
            (acc, job) => acc + job.routines.reduce((acc, routine) => acc + routine.stats.runCount, 0),
            0
        )
        return [
            {
                value: jobs.length,
                description: pluralize('Job', jobs.length),
                tooltip: 'The number of known background jobs in the system.',
            },
            {
                value: routineCount,
                description: pluralize('Routine', routineCount),
                tooltip: 'The total number of routines across all jobs.',
            },
            {
                value: hostNameCount,
                description: pluralize('Host', hostNameCount),
                tooltip: 'The total number of known hosts where jobs run.',
            },
            {
                value: routineInstanceCount,
                description: pluralize('Instance', routineInstanceCount),
                tooltip: 'The total number of routine instances across all jobs and hosts.',
            },
            {
                value: recentRunCount,
                description: pluralize('Recent run', recentRunCount),
                tooltip: 'The total number of runs tracked.',
            },
            {
                value: recentRunErrors,
                description: pluralize('Recent error', recentRunErrors),
                color: recentRunErrors > 0 ? 'var(--red)' : undefined,
                tooltip: 'The total number of errors across all runs across all routine instances.',
            },
            {
                value: runsForStats,
                description: 'Runs for stats',
                position: 'right',
                tooltip: 'The total number of runs used for calculating the stats below.',
            },
        ]
    }, [jobs, hostNameCount])

    return legends && <ValueLegendList className="mb-3" items={legends} />
}

const RoutineComponent: React.FunctionComponent<{ routine: BackgroundJob['routines'][0] }> = ({ routine }) => {
    const commonHostName = routine.recentRuns.reduce<string | undefined | null>(
        (hostName, run) => (hostName !== undefined ? run.hostName : run.hostName === hostName ? hostName : null),
        undefined
    )
    const routineIcon =
        routine.type === 'PERIODIC' ? (
            <Icon aria-hidden={true} svgPath={mdiCached} />
        ) : routine.type === 'PERIODIC_WITH_METRICS' ? (
            <Icon aria-hidden={true} svgPath={mdiNumeric} />
        ) : routine.type === 'DB_BACKED_WORKER' ? (
            <Icon aria-hidden={true} svgPath={mdiDatabase} />
        ) : routine.type === 'CUSTOM' ? (
            <Icon aria-hidden={true} svgPath={mdiShape} />
        ) : (
            <Icon aria-hidden={true} svgPath={mdiHelp} />
        )

    const roughInterval = formatDurationStructured(routine.intervalMs)[0] || { amount: 0, unit: 'millisecond' }
    const intervalUnitToColor: { [key: StructuredDuration['unit']]: string } = {
        millisecond: 'var(--orange)',
        second: 'var(--yellow)',
        minute: 'var(--green)',
        hour: 'var(--teal)',
        day: 'var(--cyan)',
    }
    const intervalColor = intervalUnitToColor[roughInterval.unit] || 'var(--gray)'

    const recentRunsTooltipContent = (
        <div>
            {commonHostName ? <Text className="mb-0">All on “{commonHostName}”:</Text> : ''}
            <ul className="pl-4">
                {routine.recentRuns.map(run => (
                    <li key={run.at}>
                        <Text className="mb-0">
                            {run.error ? <Icon aria-hidden={true} svgPath={mdiAlert} className="text-danger" /> : ''}{' '}
                            <Timestamp date={new Date(run.at)} noAbout={true} />
                            {commonHostName
                                ? ''
                                : `On host
                                                    called “${run.hostName}”,`}{' '}
                            for{' '}
                            <span className={getRunDurationTextClass(run.durationMs, routine.intervalMs)}>
                                {run.durationMs}ms
                            </span>
                            .{run.error ? ` Error: ${run.error.message}` : ''}
                        </Text>
                    </li>
                ))}
            </ul>
        </div>
    )
    const recentRunsWithErrors = routine.recentRuns.filter(run => run.error)
    const slowestRecentRunDuration = Math.max(...routine.recentRuns.map(run => run.durationMs))

    return (
        <div className={styles.routine}>
            <div>
                <ValueLegendItem
                    value={routine.intervalMs ? roughInterval.amount : '–'}
                    description={routine.intervalMs ? roughInterval.unit : 'Non-periodic'}
                    color={routine.intervalMs ? intervalColor : 'var(--gray)'}
                    tooltip={
                        routine.intervalMs
                            ? `Runs every ${formatDurationLong(routine.intervalMs)}`
                            : 'This routine is not periodic.'
                    }
                    className={styles.legendItem}
                />
            </div>
            <div className={styles.nameAndDescription}>
                <Tooltip content={routine.type.toLowerCase().replace(/_/g, ' ')} placement="top">
                    {routineIcon}
                </Tooltip>
                <Text className="mb-0 ml-2">
                    <strong>{routine.name}</strong>
                </Text>
                <div />
                <Text className="mb-0 ml-2">{routine.description}</Text>
            </div>
            <div>
                <ValueLegendItem
                    value={routine.recentRuns.length}
                    description={pluralize('recent\nrun', routine.recentRuns.length)}
                    tooltip={recentRunsWithErrors.length < routine.recentRuns.length ? recentRunsTooltipContent : ''}
                    className={styles.legendItem}
                    color={getRunDurationColor(slowestRecentRunDuration, routine.intervalMs)}
                />
            </div>
            <div>
                <ValueLegendItem
                    value={recentRunsWithErrors.length}
                    description={pluralize('recent\nerror', routine.recentRuns.length)}
                    color={recentRunsWithErrors.length ? 'var(--danger)' : 'var(--success)'}
                    tooltip={recentRunsWithErrors.length ? recentRunsTooltipContent : ''}
                    className={styles.legendItem}
                />
            </div>
            <div className="d-flex flex-column">
                {routine.stats.since ? (
                    <>
                        <div>
                            <Text className="mb-0">Fastest / avg / slowest run:</Text>
                            <Text className="mb-0">
                                <strong
                                    className={getRunDurationTextClass(routine.stats.minDurationMs, routine.intervalMs)}
                                >
                                    {routine.stats.minDurationMs}ms
                                </strong>{' '}
                                /{' '}
                                <strong
                                    className={getRunDurationTextClass(routine.stats.avgDurationMs, routine.intervalMs)}
                                >
                                    {routine.stats.avgDurationMs}ms
                                </strong>{' '}
                                /{' '}
                                <strong
                                    className={getRunDurationTextClass(routine.stats.maxDurationMs, routine.intervalMs)}
                                >
                                    {routine.stats.maxDurationMs}ms
                                </strong>
                            </Text>
                            <Text className="mb-0">
                                <span className={routine.stats.errorCount ? 'text-danger' : 'text-success'}>
                                    <strong>{routine.stats.errorCount}</strong>{' '}
                                    {pluralize('error', routine.stats.errorCount)}
                                </span>{' '}
                                in <strong>{routine.stats.runCount}</strong> {pluralize('run', routine.stats.runCount)}
                            </Text>
                            <Text className="mb-0">
                                Since <Timestamp date={new Date(routine.stats.since)} noAbout={true} />
                            </Text>
                        </div>
                    </>
                ) : (
                    'No stats.'
                )}
            </div>
        </div>
    )
}

function isRoutineProblematic(routine: BackgroundJob['routines'][0]): boolean {
    return (
        routine.stats.errorCount > 0 ||
        routine.recentRuns.some(
            run => run.error || categorizeRunDuration(run.durationMs, routine.intervalMs) !== 'short'
        ) ||
        categorizeRunDuration(routine.stats.minDurationMs, routine.intervalMs) !== 'short' ||
        categorizeRunDuration(routine.stats.avgDurationMs, routine.intervalMs) !== 'short' ||
        categorizeRunDuration(routine.stats.maxDurationMs, routine.intervalMs) !== 'short'
    )
}

// Contains some magic numbers
function categorizeRunDuration(durationMs: number, routineIntervalMs: number): RunLengthCategory {
    if (!routineIntervalMs) {
        return durationMs > 5000 ? 'long' : 'short'
    }
    if (durationMs > routineIntervalMs * 0.7) {
        return 'dangerous'
    }
    // Uses both a relative and an absolute filter
    if (durationMs > routineIntervalMs * 0.1 || durationMs > 5000) {
        return 'long'
    }
    return 'short'
}

function getRunDurationTextClass(durationMs: number, routineIntervalMs: number): string {
    const category = categorizeRunDuration(durationMs, routineIntervalMs)
    if (category === 'dangerous') {
        return 'text-danger'
    }

    if (category === 'long') {
        return 'text-warning'
    }

    return 'text-success'
}

function getRunDurationColor(durationMs: number, routineIntervalMs: number): string {
    const category = categorizeRunDuration(durationMs, routineIntervalMs)
    if (category === 'dangerous') {
        return 'var(--red)'
    }

    if (category === 'long') {
        return 'var(--yellow)'
    }

    return 'var(--green)'
}
