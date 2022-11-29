import React, { useEffect } from 'react'

import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'

export interface SiteAdminWorkersPageProps extends RouteComponentProps, TelemetryProps {
    now?: () => Date
}

export const SiteAdminWorkersPage: React.FunctionComponent<React.PropsWithChildren<SiteAdminWorkersPageProps>> = ({
    history,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminWorkers')
    }, [telemetryService])

    return (
        <div className="site-admin-workers-page">
            <PageTitle title="Workers - Admin" />
            <PageHeader
                path={[{ text: 'Workers' }]}
                headingElement="h2"
                description={<>This is the place where we're going to list stuff about workers. </>}
                className="mb-3"
            />
        </div>
    )
}
