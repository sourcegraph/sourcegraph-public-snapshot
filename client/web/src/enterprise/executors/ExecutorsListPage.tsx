import { useApolloClient } from '@apollo/client'
import CheckboxBlankCircleIcon from 'mdi-react/CheckboxBlankCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { FunctionComponent, useCallback, useEffect, useMemo } from 'react'
import { RouteComponentProps, useHistory } from 'react-router'
import { Subject } from 'rxjs'

import { Collapsible } from '@sourcegraph/web/src/components/Collapsible'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '@sourcegraph/web/src/components/FilteredConnection'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Badge, Container, PageHeader } from '@sourcegraph/wildcard'

import { ExecutorFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { queryExecutors as defaultQueryExecutors } from './useExecutors'

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

export interface ExecutorsListPageProps extends RouteComponentProps<{}> {
    queryExecutors?: typeof defaultQueryExecutors
}

export const ExecutorsListPage: FunctionComponent<ExecutorsListPageProps> = ({
    queryExecutors = defaultQueryExecutors,
    ...props
}) => {
    useEffect(() => eventLogger.logViewEvent('ExecutorsList'))

    const history = useHistory()

    const apolloClient = useApolloClient()
    const queryExecutorsCallback = useCallback(
        (args: FilteredConnectionQueryArguments) => queryExecutors(args, apolloClient),
        [queryExecutors, apolloClient]
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

            <Container className="mb-3">
                <h3>Setting up executors</h3>
                <p className="mb-0">
                    Executors enable{' '}
                    <a href="https://docs.sourcegraph.com/code_intelligence/explanations/auto_indexing" rel="noopener">
                        auto-indexing for Code Intelligence
                    </a>{' '}
                    and{' '}
                    <a href="https://docs.sourcegraph.com/batch_changes/explanations/server_side" rel="noopener">
                        server-side Batch Changes
                    </a>
                    . In order to use those features,{' '}
                    <a href="https://docs.sourcegraph.com/admin/deploy_executors" rel="noopener">
                        set them up
                    </a>
                    .
                </p>
            </Container>
            <Container>
                <FilteredConnection<ExecutorFields, {}>
                    listComponent="ul"
                    listClassName="list-group mb-2"
                    showMoreClassName="mb-0"
                    noun="executor"
                    pluralNoun="executors"
                    querySubject={querySubject}
                    nodeComponent={ExecutorNode}
                    nodeComponentProps={{}}
                    queryConnection={queryExecutorsCallback}
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

export const ExecutorNode: FunctionComponent<ExecutorNodeProps> = ({ node }) => (
    <li className="list-group-item">
        <Collapsible
            wholeTitleClickable={false}
            titleClassName="flex-grow-1"
            title={
                <div className="d-flex justify-content-between">
                    <div>
                        <h4 className="mb-0">
                            {node.active ? (
                                <CheckboxBlankCircleIcon className="icon-inline text-success mr-2" />
                            ) : (
                                <CheckboxBlankCircleIcon
                                    className="icon-inline text-warning mr-2"
                                    data-tooltip="This executor missed at least three heartbeats."
                                />
                            )}
                            {node.hostname}{' '}
                            <Badge
                                variant="secondary"
                                tooltip={`The executor is configured to pull data from the queue "${node.queueName}"`}
                            >
                                {node.queueName}
                            </Badge>
                        </h4>
                    </div>
                    <span>
                        last seen <Timestamp date={node.lastSeenAt} />
                    </span>
                </div>
            }
        >
            <dl className="mt-2 mb-0">
                <div className="d-flex w-100">
                    <div className="flex-grow-1">
                        <dt>OS</dt>
                        <dd>
                            <TelemetryData data={node.os} />
                        </dd>

                        <dt>Architecture</dt>
                        <dd>
                            <TelemetryData data={node.architecture} />
                        </dd>

                        <dt>Executor version</dt>
                        <dd>
                            <TelemetryData data={node.executorVersion} />
                        </dd>

                        <dt>Docker version</dt>
                        <dd>
                            <TelemetryData data={node.dockerVersion} />
                        </dd>
                    </div>
                    <div className="flex-grow-1">
                        <dt>Git version</dt>
                        <dd>
                            <TelemetryData data={node.gitVersion} />
                        </dd>

                        <dt>Ignite version</dt>
                        <dd>
                            <TelemetryData data={node.igniteVersion} />
                        </dd>

                        <dt>src-cli version</dt>
                        <dd>
                            <TelemetryData data={node.srcCliVersion} />
                        </dd>

                        <dt>First seen at</dt>
                        <dd>
                            <Timestamp date={node.firstSeenAt} />
                        </dd>
                    </div>
                </div>
            </dl>
        </Collapsible>
    </li>
)

export const NoExecutors: React.FunctionComponent = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1">
        <MapSearchIcon className="mb-2" />
        <br />
        No executors found.
    </p>
)

const TelemetryData: React.FunctionComponent<{ data: string }> = ({ data }) => {
    if (data) {
        return <>{data}</>
    }
    return <>n/a</>
}
