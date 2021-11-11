import { useApolloClient } from '@apollo/client'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
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
                    hideSearch={true}
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
// TODO: add additional data
export const ExecutorNode: FunctionComponent<ExecutorNodeProps> = ({ node }) => (
    <div className="mb-2">
        <table>
            <tr>
                <td>ID</td>
                <td>{node.id}</td>
            </tr>
            <tr>
                <td>Hostname</td>
                <td>{node.hostname}</td>
            </tr>
            <tr>
                <td>Last seen at</td>
                <td>
                    <Timestamp date={node.lastSeenAt} />
                </td>
            </tr>
        </table>
    </div>
)

export const NoExecutors: React.FunctionComponent = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1" data-testid="summary">
        <MapSearchIcon className="mb-2" />
        <br />
        No executors yet.
    </p>
)
