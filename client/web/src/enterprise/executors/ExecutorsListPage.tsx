import React, { FunctionComponent, useCallback, useEffect, useState } from 'react'

import { mdiCheckboxBlankCircle, mdiMapSearch } from '@mdi/js'
import { RouteComponentProps, useHistory } from 'react-router'

import { createAggregateError } from '@sourcegraph/common'
import { Badge, Container, Link, PageHeader, Icon, H3, H4, Text, Tooltip, useDebounce } from '@sourcegraph/wildcard'

import { Collapsible } from '../../components/Collapsible'
import { FilteredConnectionFilter } from '../../components/FilteredConnection'
import { useConnection } from '../../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionError,
    ConnectionForm,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../components/FilteredConnection/ui'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { ExecutorFields, ExecutorsResult, ExecutorsVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { EXECUTORS } from './backend'

const filters: FilteredConnectionFilter[] = [
    {
        id: 'state',
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

export interface ExecutorsListPageProps extends RouteComponentProps<{}> {}

export const ExecutorsListPage: FunctionComponent<React.PropsWithChildren<ExecutorsListPageProps>> = () => {
    useEffect(() => eventLogger.logPageView('ExecutorsList'))

    const history = useHistory()

    const getSearchParameter = useCallback((name: string) => new URLSearchParams(history.location.search).get(name), [
        history,
    ])

    const setSearchParameter = useCallback(
        (name: string, value: string) => {
            const parameters = new URLSearchParams(history.location.search)
            if (value !== '') {
                parameters.set(name, value)
            } else {
                parameters.delete(name)
            }

            const parameterString = parameters.toString()
            if (history.location.search !== parameterString) {
                history.replace({ ...history.location, search: parameterString })
            }
        },
        [history]
    )

    const [searchValue, setSearchValue] = useState(getSearchParameter('query') ?? '')
    const query = useDebounce(searchValue, 200)

    const [state, setState] = useState<'all' | 'active'>(() =>
        getSearchParameter('state') === 'active' ? 'active' : 'all'
    )

    const DEFAULT_VISIBLE = 20

    const { connection, loading, error, hasNextPage, fetchMore } = useConnection<
        ExecutorsResult,
        ExecutorsVariables,
        ExecutorFields
    >({
        query: EXECUTORS,
        variables: { first: DEFAULT_VISIBLE, after: null, query, active: state === 'active' },
        getConnection: ({ data, errors }) => {
            if (!data || !data.executors) {
                throw createAggregateError(errors)
            }
            return data.executors
        },
        options: {
            fetchPolicy: 'network-only',
            nextFetchPolicy: 'cache-first',
            pollInterval: 5000,
            useURL: true,
        },
    })

    const onInputChange = useCallback(
        (value: string) => {
            setSearchValue(value)
            setSearchParameter('query', value)
        },
        [setSearchParameter]
    )

    const onValueSelect = useCallback(
        (value: string) => {
            setState(value === 'active' ? 'active' : 'all')
            setSearchParameter('state', value)
        },
        [setSearchParameter]
    )

    const summary = connection && connection.nodes.length > 0 && (
        <ConnectionSummary
            connection={connection}
            first={DEFAULT_VISIBLE}
            noun="executor"
            pluralNoun="executors"
            hasNextPage={hasNextPage}
            noSummaryIfAllNodesVisible={true}
        />
    )

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
                    <Link to="/help/code_intelligence/explanations/auto_indexing" rel="noopener">
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
                <ConnectionForm
                    filters={filters}
                    inputPlaceholder="Search executors..."
                    inputValue={searchValue}
                    onInputChange={event => onInputChange(event.target.value)}
                    onValueSelect={(_filter, value) => onValueSelect(value.value)}
                    values={new Map([['state', { value: state, label: 'ignored', args: {} }]])}
                />
                <SummaryContainer>{summary}</SummaryContainer>
                {error && <ConnectionError errors={[error.message]} />}
                {loading && !connection && <ConnectionLoading />}
                {connection &&
                    (connection.nodes.length > 0 ? (
                        <ConnectionList>
                            {connection.nodes.map(node => (
                                <ExecutorNode key={node.id} node={node} />
                            ))}
                        </ConnectionList>
                    ) : (
                        <NoExecutors />
                    ))}
                <SummaryContainer>
                    {summary}
                    {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                </SummaryContainer>
            </Container>
        </>
    )
}

export interface ExecutorNodeProps {
    node: ExecutorFields
}

export const ExecutorNode: FunctionComponent<React.PropsWithChildren<ExecutorNodeProps>> = ({ node }) => (
    <li className="list-group-item">
        <Collapsible
            wholeTitleClickable={false}
            titleClassName="flex-grow-1"
            title={
                <div className="d-flex justify-content-between">
                    <div>
                        <H4 className="mb-0">
                            {node.active ? (
                                <Icon
                                    aria-hidden={true}
                                    className="text-success mr-2"
                                    svgPath={mdiCheckboxBlankCircle}
                                />
                            ) : (
                                <Tooltip content="This executor missed at least three heartbeats.">
                                    <Icon
                                        aria-label="This executor missed at least three heartbeats."
                                        className="text-warning mr-2"
                                        svgPath={mdiCheckboxBlankCircle}
                                    />
                                </Tooltip>
                            )}
                            {node.hostname}{' '}
                            <Badge
                                variant="secondary"
                                tooltip={`The executor is configured to pull data from the queue "${node.queueName}"`}
                            >
                                {node.queueName}
                            </Badge>
                        </H4>
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

export const NoExecutors: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <br />
        No executors found.
    </Text>
)

const TelemetryData: React.FunctionComponent<React.PropsWithChildren<{ data: string }>> = ({ data }) => {
    if (data) {
        return <>{data}</>
    }
    return <>n/a</>
}
