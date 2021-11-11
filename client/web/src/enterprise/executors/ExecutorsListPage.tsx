import React, { FunctionComponent, useEffect } from 'react'
import { RouteComponentProps } from 'react-router'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps, TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { useExecutors } from './useExecutors'

export interface ExecutorsListPageProps extends RouteComponentProps<{}>, TelemetryProps {
    telemetryService: TelemetryService
}

export const ExecutorsListPage: FunctionComponent<ExecutorsListPageProps> = ({ telemetryService }) => {
    useEffect(() => telemetryService.logViewEvent('ExecutorsListPage'), [telemetryService])

    const { executors, loading, error } = useExecutors()

    if (loading) {
        return <LoadingSpinner className="icon-inline" />
    }

    if (error) {
        return <ErrorAlert prefix="Error fetching executors" error={error} />
    }

    return (
        <>
            <PageTitle title="Executor instances" />
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>Executor instances</>,
                    },
                ]}
                description="The executor instances attached to your Sourcegraph instance."
                className="mb-3"
            />

            <Container>
                <ul>
                    {executors.map(executor => (
                        <li key={executor.id}>
                            <div>
                                <p>Hostname: {executor.hostname}</p>
                                <p>Last seen at: {executor.lastSeenAt}</p>
                            </div>
                        </li>
                    ))}
                </ul>
            </Container>
        </>
    )
}
