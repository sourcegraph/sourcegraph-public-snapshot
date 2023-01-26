import React, { useEffect } from 'react'

import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'

export interface SiteAdminRolesPageProps extends RouteComponentProps, TelemetryProps {}

export const SiteAdminRolesPage: React.FunctionComponent<React.PropsWithChildren<SiteAdminRolesPageProps>> = ({
    history,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminRoles')
    }, [telemetryService])

    console.log('inside roles page')
    return (
        <div className="site-admin-roles-page">
            <PageTitle title="Roles - Admin" />
            <PageHeader
                path={[{ text: 'Roles' }]}
                headingElement="h2"
                description={
                    <>
                        This is the log of recent external requests sent by the Sourcegraph instance. Handy for seeing
                        what's happening between Sourcegraph and other services.{' '}
                        <strong>The list updates every five seconds.</strong>
                    </>
                }
                className="mb-3"
            />
        </div>
    )
}
