import React, { useEffect } from 'react'

import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, ErrorAlert, H3, H4, Text, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'
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

// {color: green/red if running} {type} {Job name} {running instance count, with a list as tooltip} {Uptime (relative) OR Min: Avg: Max: relative uptime across instances}
// - {color: green/red if running} {Routine name} {Description icon} {running instance count, with a list as tooltip} {stats}
// {color: green/red if successful} {Last run #} {at} {durationMs} {success / error}
const JobList: React.FunctionComponent<{ jobs: BackgroundJobs[] }> = ({ jobs }) => (
    <ul>
        {jobs.map(job => (
            <li key={job.name}>
                <H3>{job.name}</H3>
                {job.routines.map(routine => (
                    <div key={routine.name}>
                        <H4>{routine.name}</H4>
                        <Text>{routine.description}</Text>
                    </div>
                ))}
            </li>
        ))}
    </ul>
)
