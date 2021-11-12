import { useApolloClient } from '@apollo/client'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { FunctionComponent, useCallback, useEffect, useMemo } from 'react'
import { RouteComponentProps, useHistory } from 'react-router'
import { Subject } from 'rxjs'

import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '@sourcegraph/web/src/components/FilteredConnection'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Container, PageHeader } from '@sourcegraph/wildcard'

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

            <Container>
                <FilteredConnection<ExecutorFields, {}>
                    listComponent="div"
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

// TODO: style
export const ExecutorNode: FunctionComponent<ExecutorNodeProps> = ({ node }) => (
    <div className="p-2">
        <hr />

        <dl>
            <dt>ID</dt>
            <dd>{node.id}</dd>
            <dt>Hostname</dt>
            <dd>{node.hostname}</dd>
            <dt>Queue Name</dt>
            <dd>{node.queueName}</dd>
            <dt>OS</dt>
            <dd>{node.os}</dd>
            <dt>Architecture</dt>
            <dd>{node.architecture}</dd>
            <dt>Executor version</dt>
            <dd>{node.executorVersion}</dd>
            <dt>src-cli version</dt>
            <dd>{node.srcCliVersion}</dd>
            <dt>Docker version</dt>
            <dd>{node.dockerVersion}</dd>
            <dt>Git version</dt>
            <dd>{node.gitVersion}</dd>
            <dt>Ignite version</dt>
            <dd>{node.igniteVersion}</dd>
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
