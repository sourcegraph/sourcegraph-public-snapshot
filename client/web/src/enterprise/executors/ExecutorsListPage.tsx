import { useApolloClient } from '@apollo/client'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { FunctionComponent, useCallback, useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import { Subject } from 'rxjs'

import { TelemetryProps, TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '@sourcegraph/web/src/components/FilteredConnection'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { ExecutorFields } from '../../graphql-operations'

import { useExecutors as defaultDoQueryExecutors } from './useExecutors'

const filters: FilteredConnectionFilter[] = [
    {
        id: 'filters',
        label: 'State',
        type: 'select',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all executors',
                args: {},
            },
            {
                label: 'Active',
                value: 'active',
                tooltip: 'Show only active executors',
                args: { active: true },
            },
        ],
    },
]

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

            <Container>
                <FilteredConnection<ExecutorFields, {}>
                    listComponent="div"
                    noun="executor"
                    pluralNoun="executors"
                    querySubject={querySubject}
                    nodeComponent={ExecutorNode}
                    nodeComponentProps={{}}
                    queryConnection={queryExecutors}
                    history={history}
                    location={props.location}
                    cursorPaging={true}
                    filters={filters}
                    emptyElement={<NoExecutors />}
                />
            </Container>
        </>
    )
}

export interface ExecutorNodeProps {
    node: ExecutorFields
}

// TODO: style
export const ExecutorNode: FunctionComponent<ExecutorNodeProps> = ({ node }) => (
    <div className="p-2">
        <hr />

        <dl>
            {/*
            <dt>ID</dt>
            <dl>{node.id}</dl>
            */}
            <dt>Hostname</dt>
            <dl>{node.hostname}</dl>
            <dt>Queue Name</dt>
            <dl>{node.queueName}</dl>

            {/*
            <dt>OS</dt>
            <dl>{node.os}</dl>
            <dt>Architecture</dt>
            <dl>{node.architecture}</dl>
            <dt>Executor version</dt>
            <dl>{node.executorVersion}</dl>
            <dt>src-cli version</dt>
            <dl>{node.srcCliVersion}</dl>
            <dt>Docker version</dt>
            <dl>{node.dockerVersion}</dl>
            <dt>Ignite version</dt>
            <dl>{node.igniteVersion}</dl>
            */}

            <dt>First seen at</dt>
            <dd>
                <Timestamp date={node.firstSeenAt} />
            </dd>

            <dt>Last seen at</dt>
            <dd>
                <Timestamp date={node.lastSeenAt} />
            </dd>
        </dl>

        <hr />
    </div>
)

export const NoExecutors: React.FunctionComponent = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1" data-testid="summary">
        <MapSearchIcon className="mb-2" />
        <br />
        No executors yet.
    </p>
)
