import React, { useEffect } from 'react'

import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'

import styles from './SiteAdminBackgroundJobsPage.module.scss'

export interface SiteAdminBackgroundJobsPageProps extends RouteComponentProps, TelemetryProps {
    now?: () => Date
}

export const SiteAdminBackgroundJobsPage: React.FunctionComponent<
    React.PropsWithChildren<SiteAdminBackgroundJobsPageProps>
> = ({ history, telemetryService }) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminBackgroundJobs')
    }, [telemetryService])

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
        </div>
    )
}
