import * as H from 'history'
import React, { useMemo } from 'react'
import { Breadcrumbs, BreadcrumbSetters, BreadcrumbsProps } from '../components/Breadcrumbs'
import { PageHeader } from '../components/PageHeader'
import VideoInputAntennaIcon from 'mdi-react/VideoInputAntennaIcon'

interface CodeMonitoringPageProps extends BreadcrumbsProps, BreadcrumbSetters {
    location: H.Location
}

export const CodeMonitoringPage: React.FunctionComponent<CodeMonitoringPageProps> = props => {
    props.useBreadcrumb(
        useMemo(
            () => ({
                key: 'Code Monitoring',
                element: <>Code Monitoring</>,
            }),
            []
        )
    )

    return (
        <div className="w-100">
            <Breadcrumbs breadcrumbs={props.breadcrumbs} location={props.location} />
            <div className="container mt-3 web-content">
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
                    actions={<></>}
                />
                <div>Hello world</div>
            </div>
        </div>
    )
}
