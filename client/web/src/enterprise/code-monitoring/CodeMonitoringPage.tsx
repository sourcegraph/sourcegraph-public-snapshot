import * as H from 'history'
import React from 'react'
import VideoInputAntennaIcon from 'mdi-react/VideoInputAntennaIcon'
import { BreadcrumbSetters, BreadcrumbsProps } from '../../components/Breadcrumbs'
import { PageHeader } from '../../components/PageHeader'
import { PageTitle } from '../../components/PageTitle'

export interface CodeMonitoringPageProps extends BreadcrumbsProps, BreadcrumbSetters {
    location: H.Location
}

export const CodeMonitoringPage: React.FunctionComponent<CodeMonitoringPageProps> = props => (
    <div className="container mt-3 web-content">
        <PageTitle title="Code Monitoring" />
        <PageHeader
            title={
                <>
                    Code Monitoring{' '}
                    <sup>
                        <span className="badge badge-info text-uppercase">Prototype</span>
                    </sup>
                </>
            }
            icon={VideoInputAntennaIcon}
        />
        <div>Hello world</div>
    </div>
)
