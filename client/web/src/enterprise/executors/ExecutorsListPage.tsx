import { useApolloClient } from '@apollo/client'
import React, { FunctionComponent, useCallback, useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import { Subject } from 'rxjs'

import { TelemetryProps, TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    FilteredConnection,
    FilteredConnectionQueryArguments,
} from '@sourcegraph/web/src/components/FilteredConnection'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { ExecutorFields } from '../../graphql-operations'

import { useExecutors as defaultDoQueryExecutors } from './useExecutors'

export interface ExecutorsListPageProps extends RouteComponentProps<{}>, TelemetryProps {
    telemetryService: TelemetryService
    doQueryExecutors?: typeof defaultDoQueryExecutors
}

export const ExecutorsListPage: FunctionComponent<ExecutorsListPageProps> = ({
    doQueryExecutors = defaultDoQueryExecutors,
    telemetryService,
    history,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('ExecutorsListPage'), [telemetryService])

    const apolloClient = useApolloClient()
    const queryExecutors = useCallback(
        (args: FilteredConnectionQueryArguments) => doQueryExecutors(args, apolloClient),
        [doQueryExecutors, apolloClient]
    )

    const querySubject = useMemo(() => new Subject<string>(), [])

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

            <Container className="mb-2">
                <FilteredConnection<ExecutorFields, {}>
                    // listComponent
                    // listClassName
                    noun="executor"
                    pluralNoun="executors"
                    querySubject={querySubject}
                    nodeComponent={ExecutorNode}
                    nodeComponentProps={{}}
                    queryConnection={queryExecutors}
                    history={history}
                    location={props.location}
                    cursorPaging={true}
                    // filters
                    // emptyElement
                />
            </Container>
        </>
    )
}

export interface ExecutorNodeProps {
    node: ExecutorFields
}

export const ExecutorNode: FunctionComponent<ExecutorNodeProps> = ({ node }) => (
    <div>
        <p>ID: {node.id}</p>
        <p>Hostname: {node.hostname}</p>
        <p>
            Last seen at: <Timestamp date={node.lastSeenAt} />
        </p>
    </div>
)
