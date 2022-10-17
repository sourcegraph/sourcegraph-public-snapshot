import React, { FunctionComponent, useCallback, useEffect, useMemo } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiCheckboxBlankCircle, mdiMapSearch } from '@mdi/js'
import { RouteComponentProps, useHistory } from 'react-router'
import { Subject } from 'rxjs'

import { Badge, Container, Link, PageHeader, Icon, H3, H4, Text, Tooltip } from '@sourcegraph/wildcard'

import { Collapsible } from '../../../components/Collapsible'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import { Timestamp } from '../../../components/time/Timestamp'
import { ExecutorFields } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'

import { ExecutorCompatibilityAlert } from './ExecutorCompatibilityAlert'
import { queryExecutors as defaultQueryExecutors } from './useExecutors'
import { ExecutorNode } from './ExecutorNode'

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

export const ExecutorsListPage: FunctionComponent<React.PropsWithChildren<ExecutorsListPageProps>> = ({
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
                <H3>Setting up executors</H3>
                <Text className="mb-0">
                    Executors enable{' '}
                    <Link to="/help/code_navigation/explanations/auto_indexing" rel="noopener">
                        auto-indexing for code navigation
                    </Link>{' '}
                    and{' '}
                    <Link to="/help/batch_changes/explanations/server_side" rel="noopener">
                        running batch changes server-side
                    </Link>
                    . In order to use those features,{' '}
                    <Link to="/help/admin/deploy_executors" rel="noopener">
                        set them up
                    </Link>
                    .
                </Text>
            </Container>
            <Container>
                <FilteredConnection<ExecutorFields>
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

export const NoExecutors: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <br />
        No executors found.
    </Text>
)
