import React, { useCallback, useEffect, useMemo, useState } from 'react'

import {
    mdiAccountHardHat,
    mdiAlert,
    mdiCached,
    mdiCheck,
    mdiClose,
    mdiDatabase,
    mdiHelp,
    mdiNumeric,
    mdiShape,
} from '@mdi/js'
import format from 'date-fns/format'
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
    Link,
    LoadingSpinner,
    PageHeader,
    Select,
    Text,
    Tooltip,
    useSessionStorage,
} from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'
import { BackgroundJobsResult, BackgroundJobsVariables, BackgroundRoutineType } from '../graphql-operations'
import { formatDurationLong } from '../util/time'

import { ValueLegendList, ValueLegendListProps } from './analytics/components/ValueLegendList'
import { BACKGROUND_JOBS, BACKGROUND_JOBS_PAGE_POLL_INTERVAL_MS } from './backend'

import styles from './SiteAdminBackgroundJobsPage.module.scss'

export interface SiteAdminBackgroundJobsPageProps extends RouteComponentProps, TelemetryProps {}

export type BackgroundJob = BackgroundJobsResult['backgroundJobs']['nodes'][0]

// "short" runs are displayed with a “success” style.
// “long” runs are displayed with a “warning” style to make sure they stand out somewhat.
// “dangerous” runs are displayed with a “danger” style to make sure they stand out even more.
type RunLengthCategory = 'short' | 'long' | 'dangerous'

// The maximum number of recent runs to fetch for each routine.
const recentRunCount = 5

export const SiteAdminBackgroundJobsPage: React.FunctionComponent<
    React.PropsWithChildren<SiteAdminBackgroundJobsPageProps>
> = ({ telemetryService }) => {
    // Log page view
    useEffect(() => {
        telemetryService.logPageView('SiteAdminBackgroundJobs')
    }, [telemetryService])

    // Data query and polling setting
    const { data, loading, error, stopPolling, startPolling } = useQuery<BackgroundJobsResult, BackgroundJobsVariables>(
        BACKGROUND_JOBS,
        {
            variables: { recentRunCount },
            pollInterval: BACKGROUND_JOBS_PAGE_POLL_INTERVAL_MS,
        }
    )
    const [polling, setPolling] = useState(true)
    const togglePolling = useCallback(() => {
        if (polling) {
            stopPolling()
        } else {
            startPolling(BACKGROUND_JOBS_PAGE_POLL_INTERVAL_MS)
        }
        setPolling(!polling)
    }, [polling, startPolling, stopPolling])

    return (
        <div>
            <PageTitle title="Background jobs - Admin" />
            <Button variant="secondary" onClick={togglePolling} className="float-right">
                {polling ? 'Pause polling' : 'Resume polling'}
            </Button>
            <PageHeader
                path={[{ text: 'Background jobs' }]}
                headingElement="h2"
                description={
                    <>
                        This page lists{' '}
                        <Link to="/help/admin/workers" target="_blank" rel="noopener noreferrer">
                            all running jobs
                        </Link>
                        , their routines, recent runs, any errors, timings, and stats.
                    </>
                }
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
                {!loading && !error && data?.backgroundJobs.nodes && <JobList jobs={data.backgroundJobs.nodes} />}
            </Container>
        </div>
    )
}

const JobList: React.FunctionComponent<{
    jobs: BackgroundJob[]
}> = ({ jobs }) => {
    const [onlyShowProblematic, setOnlyShowProblematic] = useSessionStorage(
        'site-admin.background-jobs.only-show-problematic-routines',
        false
    )

    const hostNames = useMemo(
        () =>
            jobs
                .map(job => job.routines[0]?.instances[0]?.hostName) // get the host name of the first routine
                .filter((host, index, hosts) => hosts.indexOf(host) === index) // deduplicate
                .filter(host => !!host), // remove undefined
        [jobs]
    )

    const problematicJobs = useMemo(
        () => jobs.filter(job => job.routines.some(routine => isRoutineProblematic(routine))),
        [jobs]
    )
    const jobsToDisplay = onlyShowProblematic ? problematicJobs : jobs

    return (
        <>
            <LegendList jobs={jobs} hostNameCount={hostNames.length} />
            {jobsToDisplay ? (
                <>
                    <div className={styles.tableHeader}>
                        <div>
                            <Select
                                aria-label="Filter for problematic routines"
                                onChange={value => setOnlyShowProblematic(value.target.value === 'problematic')}
                                selectClassName={styles.filterSelect}
                                defaultValue={onlyShowProblematic ? 'problematic' : 'all'}
                            >
                                <option value="all">Show all routines</option>
                                <option value="problematic">Only show problematic routines</option>
                            </Select>
                        </div>
                        <div className="text-center">Fastest / avg / slowest run (ms)</div>
                    </div>
                    <ul className="list-group list-group-flush">
                        {jobsToDisplay.map(job => {
                            const jobHostNames = [
                                ...new Set(
                                    job.routines
                                        .map(routine => routine.instances.map(instance => instance.hostName))
                                        .flat()
                                ),
                            ].sort()
                            return (
                                <li key={job.name} className="list-group-item px-0 py-2">
                                    <div className="d-flex align-items-center justify-content-between mb-2">
                                        <div className="d-flex flex-row align-items-center mb-0">
                                            <Icon aria-hidden={true} svgPath={mdiAccountHardHat} />{' '}
                                            <Text className="mb-0 ml-2">
                                                <strong>{job.name}</strong>{' '}
                                                <span className="text-muted">
                                                    (starts {job.routines.length}{' '}
                                                    {pluralize('routine', job.routines.length)}
                                                    {hostNames.length > 1
                                                        ? ` on ${jobHostNames.length} ${pluralize(
                                                              'instance',
                                                              jobHostNames.length
                                                          )}`
                                                        : ''}
                                                    )
                                                </span>
                                            </Text>
                                        </div>
                                    </div>
                                    {job.routines
                                        .filter(routine => (onlyShowProblematic ? isRoutineProblematic(routine) : true))
                                        .map(routine => (
                                            <RoutineItem routine={routine} key={routine.name} />
                                        ))}
                                </li>
                            )
                        })}
                    </ul>
                </>
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
        const recentRunErrors = jobs.reduce(
            (acc, job) =>
                acc +
                job.routines.reduce(
                    (acc, routine) => acc + routine.recentRuns.filter(run => run.errorMessage).length,
                    0
                ),
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
                value: recentRunErrors,
                description: pluralize('Recent error', recentRunErrors),
                color: recentRunErrors > 0 ? 'var(--red)' : undefined,
                tooltip: 'The total number of errors across all runs across all routine instances.',
            },
        ]
    }, [jobs, hostNameCount])

    return legends && <ValueLegendList className="mb-3" items={legends} />
}

const RoutineItem: React.FunctionComponent<{ routine: BackgroundJob['routines'][0] }> = ({ routine }) => {
    const commonHostName =
        routine.recentRuns.length === 1
            ? routine.recentRuns[0].hostName
            : routine.recentRuns.reduce<string | undefined | null>(
                  (hostName, run) =>
                      hostName !== undefined ? run.hostName : run.hostName === hostName ? hostName : null,
                  undefined
              )
    const routineIcon =
        routine.type === 'PERIODIC' ? (
            <Icon aria-hidden={true} svgPath={mdiCached} />
        ) : routine.type === 'PERIODIC_WITH_METRICS' ? (
            <Icon aria-hidden={true} svgPath={mdiNumeric} />
        ) : routine.type === 'DB_BACKED' ? (
            <Icon aria-hidden={true} svgPath={mdiDatabase} />
        ) : routine.type === 'CUSTOM' ? (
            <Icon aria-hidden={true} svgPath={mdiShape} />
        ) : (
            <Icon aria-hidden={true} svgPath={mdiHelp} />
        )

    const recentRunsReversed = [...routine.recentRuns].reverse()

    const recentRunsTooltipContent = (
        <div>
            {commonHostName ? <Text className="mb-0">All on “{commonHostName}”:</Text> : ''}
            <ul className="pl-4">
                {recentRunsReversed.map(run => (
                    <li key={run.at}>
                        <Text className="mb-0">
                            {run.errorMessage ? (
                                <Icon aria-hidden={true} svgPath={mdiAlert} className="text-danger" />
                            ) : (
                                ''
                            )}{' '}
                            <Timestamp date={new Date(run.at)} noAbout={true} />
                            {commonHostName ? '' : ` on the host called “${run.hostName}”,`} for{' '}
                            <span className={getRunDurationTextClass(run.durationMs, routine.intervalMs)}>
                                {run.durationMs}ms
                            </span>
                            .{run.errorMessage ? ` Error: ${run.errorMessage}` : ''}
                        </Text>
                    </li>
                ))}
            </ul>
        </div>
    )
    const recentRunsWithErrors = routine.recentRuns.filter(run => run.errorMessage)

    return (
        <div className={styles.routine}>
            <div className={styles.nameAndDescription}>
                <Text className="mb-1 ml-4">
                    <span className="mr-2">
                        <StartedStoppedIndicator routine={routine} />
                    </span>
                    <Tooltip content={routine.type.toLowerCase().replace(/_/g, ' ')} placement="top">
                        {routineIcon}
                    </Tooltip>
                    <span className="ml-2">
                        <strong>{routine.name}</strong>
                    </span>
                    <span className="ml-2 text-muted">{routine.description}</span>
                </Text>
                <Text className="mb-0 ml-4 text-muted">
                    {routine.intervalMs ? (
                        <>
                            {routine.type !== 'DB_BACKED' ? 'Runs ' : 'Checks queue '}every{' '}
                            <strong>{formatDurationLong(routine.intervalMs)}</strong>.{' '}
                        </>
                    ) : null}
                    {routine.recentRuns.length > 0 ? (
                        <Tooltip content={recentRunsTooltipContent}>
                            <span>
                                <strong>
                                    <span className={recentRunsWithErrors.length ? 'text-danger' : 'text-success'}>{`${
                                        recentRunsWithErrors.length
                                    } ${pluralize('error', recentRunsWithErrors.length)}`}</span>
                                </strong>
                                <span className={styles.linkColor}>*</span> in the last{' '}
                                {`${routine.recentRuns.length} ${pluralize('run', routine.recentRuns.length)}`}.{' '}
                            </span>
                        </Tooltip>
                    ) : null}
                    {routine.stats.runCount ? (
                        <>
                            <span className={routine.stats.errorCount ? 'text-danger' : 'text-success'}>
                                <strong>
                                    {routine.stats.errorCount} {pluralize('error', routine.stats.errorCount)}
                                </strong>
                            </span>{' '}
                            in <strong>{routine.stats.runCount}</strong> {pluralize('run', routine.stats.runCount)}
                            {routine.stats.since ? (
                                <>
                                    {' '}
                                    in the last{' '}
                                    <Timestamp date={new Date(routine.stats.since)} noAbout={true} noAgo={true} />.
                                </>
                            ) : null}
                        </>
                    ) : null}
                </Text>
            </div>
            <div className="text-center">
                {routine.stats.runCount ? (
                    <Tooltip content="Fastest / avg / slowest run in milliseconds">
                        <div>
                            <span className={getRunDurationTextClass(routine.stats.minDurationMs, routine.intervalMs)}>
                                {routine.stats.minDurationMs}
                            </span>{' '}
                            /{' '}
                            <span className={getRunDurationTextClass(routine.stats.avgDurationMs, routine.intervalMs)}>
                                {routine.stats.avgDurationMs}
                            </span>{' '}
                            /{' '}
                            <span className={getRunDurationTextClass(routine.stats.maxDurationMs, routine.intervalMs)}>
                                {routine.stats.maxDurationMs}
                            </span>
                        </div>
                    </Tooltip>
                ) : (
                    <span className="text-muted">No stats.</span>
                )}
            </div>
        </div>
    )
}

const StartedStoppedIndicator: React.FunctionComponent<{ routine: BackgroundJob['routines'][0] }> = ({ routine }) => {
    const latestStartDateString = routine.instances.reduce(
        (mostRecent, instance) =>
            instance.lastStartedAt && (!mostRecent || instance.lastStartedAt > mostRecent)
                ? instance.lastStartedAt
                : mostRecent,
        ''
    )
    const earliestStopDateString = routine.instances.reduce(
        (earliest, instance) =>
            instance.lastStoppedAt && (!earliest || instance.lastStoppedAt < earliest)
                ? instance.lastStoppedAt
                : earliest,
        ''
    )
    const lastRecentRunDate =
        routine.recentRuns.length && new Date(routine.recentRuns[routine.recentRuns.length - 1].at)
    const isStopped = earliestStopDateString && earliestStopDateString >= latestStartDateString
    const isUnseenInAWhile =
        routine.intervalMs &&
        routine.type !== BackgroundRoutineType.DB_BACKED &&
        (!lastRecentRunDate ||
            lastRecentRunDate.getTime() + routine.intervalMs + routine.stats.maxDurationMs + 5000 <= Date.now())

    const tooltip = isStopped
        ? `This routine is currently stopped.
Started at ${format(new Date(latestStartDateString), 'yyyy-MM-dd HH:mm:ss')},
Stopped at: ${format(new Date(earliestStopDateString), 'yyyy-MM-dd HH:mm:ss')}`
        : isUnseenInAWhile
        ? lastRecentRunDate
            ? `This routine has not been seen in a while. It should've run at ${format(
                  new Date(lastRecentRunDate.getTime() + (routine.intervalMs || 0)),
                  'yyyy-MM-dd HH:mm:ss'
              )}.`
            : 'This routine was started but it has never been seen running.'
        : `This routine is currently started.${
              lastRecentRunDate ? `\nLast seen running at ${format(lastRecentRunDate, 'yyyy-MM-dd HH:mm:ss')}.` : ''
          }`

    return isStopped || isUnseenInAWhile ? (
        <Tooltip content={tooltip}>
            <Icon aria-label="stopped" svgPath={mdiClose} className="text-danger" />
        </Tooltip>
    ) : (
        <Tooltip content={tooltip}>
            <Icon aria-label="started" svgPath={mdiCheck} className="text-success" />
        </Tooltip>
    )
}

function isRoutineProblematic(routine: BackgroundJob['routines'][0]): boolean {
    return (
        routine.stats.errorCount > 0 ||
        routine.recentRuns.some(
            run => run.errorMessage || categorizeRunDuration(run.durationMs, routine.intervalMs) !== 'short'
        ) ||
        categorizeRunDuration(routine.stats.minDurationMs, routine.intervalMs) !== 'short' ||
        categorizeRunDuration(routine.stats.avgDurationMs, routine.intervalMs) !== 'short' ||
        categorizeRunDuration(routine.stats.maxDurationMs, routine.intervalMs) !== 'short'
    )
}

// Contains some magic numbers
function categorizeRunDuration(durationMs: number, routineIntervalMs: number | null): RunLengthCategory {
    if (!routineIntervalMs) {
        return durationMs > 5000 ? 'long' : 'short'
    }
    if (durationMs > routineIntervalMs * 0.7) {
        return 'dangerous'
    }
    // Uses both a relative and an absolute filter:
    // If the run is more than 10% longer than the interval, it's long, except for intervals of 1s or less where 500ms+ is long.
    // Also, any run longer than 5s is long.
    if (durationMs > routineIntervalMs * (durationMs > 1000 ? 0.1 : 0.5) || durationMs > 5000) {
        return 'long'
    }
    return 'short'
}

function getRunDurationTextClass(durationMs: number, routineIntervalMs: number | null): string {
    const category = categorizeRunDuration(durationMs, routineIntervalMs)
    if (category === 'dangerous') {
        return 'text-danger'
    }

    if (category === 'long') {
        return 'text-warning'
    }

    return 'text-success'
}
