import * as H from 'history'
import VideoInputAntennaIcon from 'mdi-react/VideoInputAntennaIcon'
import React, { useMemo } from 'react'
import { AuthenticatedUser } from '../../auth'
import { BreadcrumbSetters, BreadcrumbsProps } from '../../components/Breadcrumbs'
import { PageHeader } from '../../components/PageHeader'
import { PageTitle } from '../../components/PageTitle'
import { CodeMonitorForm } from './components/CodeMonitorForm'

export interface CreateCodeMonitorPageProps extends BreadcrumbsProps, BreadcrumbSetters {
    location: H.Location
    authenticatedUser: AuthenticatedUser
}

export const CreateCodeMonitorPage: React.FunctionComponent<CreateCodeMonitorPageProps> = props => {
    props.useBreadcrumb(
        useMemo(
            () => ({
                key: 'Create Code Monitor',
                element: <>Create new code monitor</>,
            }),
            []
        )
    )

    return (
        <div className="container mt-3 web-content">
            <PageTitle title="Create new code monitor" />
            <PageHeader title="Create new code monitor" icon={VideoInputAntennaIcon} />
            Code monitors watch your code for specific triggers and run actions in response.{' '}
            <a href="" target="_blank" rel="noopener">
                {/* TODO: populate link */}
                Learn more
            </a>
            <CodeMonitorForm {...props} />
        </div>
    )
}
