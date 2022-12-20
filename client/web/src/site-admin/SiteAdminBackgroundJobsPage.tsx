import React, { useEffect } from 'react'

import { mdiAccountHardHat, mdiAlert, mdiCached, mdiNumeric, mdiRunFast, mdiText } from '@mdi/js'
import { RouteComponentProps } from 'react-router'

import { pluralize } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, ErrorAlert, Icon, LoadingSpinner, PageHeader, Text, Tooltip } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'
import { Timestamp } from '../components/time/Timestamp'
import { BackgroundJobsResult, BackgroundJobsVariables } from '../graphql-operations'

import { BACKGROUND_JOBS, BACKGROUND_JOBS_PAGE_POLL_INTERVAL_MS } from './backend'

import styles from './SiteAdminBackgroundJobsPage.module.scss'

export interface SiteAdminBackgroundJobsPageProps extends RouteComponentProps, TelemetryProps {}

export type BackgroundJobs = BackgroundJobsResult['backgroundJobs']['nodes'][0]

export const SiteAdminBackgroundJobsPage: React.FunctionComponent<
    React.PropsWithChildren<SiteAdminBackgroundJobsPageProps>
> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminBackgroundJobs')
    }, [telemetryService])

    const { data, loading, error } = useQuery<BackgroundJobsResult, BackgroundJobsVariables>(BACKGROUND_JOBS, {
        variables: { recentRunCount: 5 },
        pollInterval: BACKGROUND_JOBS_PAGE_POLL_INTERVAL_MS,
    })

    return (
        <div className={styles.page}>
            <PageTitle title="Background jobs - Admin" />
            <PageHeader
                path={[{ text: 'Background jobs' }]}
                headingElement="h2"
                description={
                    <>This is the place where we're going to list stuff about background jobs and routines. </>
                }
                className="mb-3"
            />
            <Container className="mb-3">
                {error && !loading && <ErrorAlert error={error} />}
                {loading && !error && <LoadingSpinner />}
                {!loading && !error && data?.backgroundJobs.nodes && <JobList jobs={data.backgroundJobs.nodes} />}
            </Container>
        </div>
    )
}

// {color: green/red if running} {type} {Job name} {routine count} {Uptime (relative) OR Min: Avg: Max: relative uptime across instances}
// - {color: green/red if running} {Routine name} {Description icon} {running instance count, with a list as tooltip} {stats}
// {color: green/red if successful} {Last run #} {at} {durationMs} {success / error}
const JobList: React.FunctionComponent<{ jobs: BackgroundJobs[] }> = ({ jobs }) => (
    <ul>
        {jobs.map(job => {
            const hostNames = [
                ...new Set(job.routines.map(routine => routine.instances.map(instance => instance.hostName)).flat()),
            ].sort()
            return (
                <li key={job.name} className="list-group-item px-0 py-2">
                    <div className="d-flex align-items-center justify-content-between">
                        <div className="d-flex flex-row">
                            <Icon aria-hidden={true} svgPath={mdiAccountHardHat} /> <Text>{job.name}</Text>
                        </div>
                        <div>
                            <Text className="mb-0">
                                {hostNames.length} {pluralize('running instance', hostNames.length)}
                            </Text>
                            <Text className="mb-0">
                                {job.routines.length} {pluralize('started routine', job.routines.length)}
                            </Text>
                        </div>
                    </div>
                    <ul>
                        {job.routines.map(routine => {
                            const commonHostName = routine.recentRuns.reduce<string | undefined | null>(
                                (hostName, run) =>
                                    hostName !== undefined ? run.hostName : run.hostName === hostName ? hostName : null,
                                undefined
                            )
                            const routineIcon =
                                routine.type === 'PERIODIC' ? (
                                    <Icon aria-hidden={true} svgPath={mdiCached} />
                                ) : routine.type === 'PERIODIC_WITH_METRICS' ? (
                                    <>
                                        <Icon aria-hidden={true} svgPath={mdiCached} />
                                        <Icon aria-hidden={true} svgPath={mdiNumeric} />
                                    </>
                                ) : (
                                    <>Unknown</>
                                )

                            return (
                                <li key={routine.name}>
                                    <div className="d-flex flex-row">
                                        <div>
                                            <Tooltip content={routine.type.toLowerCase()} placement="top">
                                                {routineIcon}
                                            </Tooltip>
                                        </div>
                                        <Text className="mb-0">{routine.name}</Text>
                                    </div>
                                    <div className="d-flex flex-row">
                                        <Icon aria-hidden={true} svgPath={mdiText} />
                                        <div>Description: {routine.description}</div>
                                    </div>
                                    <div>
                                        <div className="d-flex flex-row">
                                            <Icon aria-hidden={true} svgPath={mdiRunFast} />
                                            <Tooltip
                                                content={
                                                    <ul>
                                                        {routine.recentRuns.map(run => (
                                                            <li key={run.at}>
                                                                <div className="d-flex flex-row">
                                                                    <Text className="mb-0">
                                                                        {run.error ? (
                                                                            <Icon
                                                                                aria-hidden={true}
                                                                                svgPath={mdiAlert}
                                                                            />
                                                                        ) : (
                                                                            ''
                                                                        )}{' '}
                                                                        <Timestamp
                                                                            date={new Date(run.at)}
                                                                            noAbout={true}
                                                                        />
                                                                        {commonHostName
                                                                            ? ''
                                                                            : `On host
                                                            called “${run.hostName}”,`}{' '}
                                                                        for {run.durationMs}ms.
                                                                        {run.error
                                                                            ? ` Error: ${run.error.message}`
                                                                            : ''}
                                                                    </Text>
                                                                </div>
                                                            </li>
                                                        ))}
                                                    </ul>
                                                }
                                                placement="bottom"
                                            >
                                                <Text className="mb-0">
                                                    {routine.recentRuns.length}{' '}
                                                    {pluralize('recent run', routine.recentRuns.length)}
                                                    {commonHostName ? `, all on “${commonHostName}”` : ''}:
                                                </Text>
                                            </Tooltip>
                                        </div>
                                    </div>
                                    <Text className="mb-0">
                                        📊{' '}
                                        {routine.stats.since ? (
                                            <>
                                                Ran{' '}
                                                <strong>
                                                    {routine.stats.runCount !== 1
                                                        ? `${routine.stats.runCount} times`
                                                        : 'once'}
                                                </strong>{' '}
                                                since{' '}
                                                <strong>
                                                    <Timestamp date={new Date(routine.stats.since)} noAbout={true} />
                                                </strong>{' '}
                                                with {routine.stats.errorCount}{' '}
                                                {pluralize('error', routine.stats.errorCount)}. Fastest run:{' '}
                                                <strong>{routine.stats.minDurationMs}ms</strong>, slowest run:{' '}
                                                <strong>{routine.stats.maxDurationMs}ms</strong>, average run:{' '}
                                                <strong>{routine.stats.avgDurationMs}ms</strong>.
                                            </>
                                        ) : (
                                            'No runs recorded, so no stats.'
                                        )}
                                    </Text>
                                </li>
                            )
                        })}
                    </ul>
                </li>
            )
        })}
    </ul>
)
