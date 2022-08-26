import React, { FunctionComponent, useCallback, useEffect, useMemo } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiCheckboxBlankCircle, mdiMapSearch } from '@mdi/js'
import { parse, isAfter } from 'date-fns'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { RouteComponentProps, useHistory } from 'react-router'
import { Subject } from 'rxjs'
import semver from 'semver'

import { useQuery } from '@sourcegraph/http-client'
import {
    Badge,
    LoadingSpinner,
    Container,
    Link,
    PageHeader,
    Icon,
    H3,
    H4,
    Text,
    Tooltip,
    Alert,
} from '@sourcegraph/wildcard'

import { Collapsible } from '../../components/Collapsible'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../components/FilteredConnection'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { ExecutorFields, GetSourcegraphVersionResult, GetSourcegraphVersionVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { GET_SOURCEGRAPH_VERSION } from './backend'
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

    const { data, loading, error } = useQuery<GetSourcegraphVersionResult, GetSourcegraphVersionVariables>(
        GET_SOURCEGRAPH_VERSION,
        {
            fetchPolicy: 'cache-and-network',
        }
    )

    const querySubject = useMemo(() => new Subject<string>(), [])

    if (error || !data?.site) {
        const title = error ? String(error) : 'Unable to fetch sourcegraph version.'
        return <HeroPage icon={AlertCircleIcon} title={title} />
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
                {loading ? (
                    <LoadingSpinner />
                ) : (
                    <FilteredConnection<
                        ExecutorFields,
                        {
                            sourcegraphVersion: string
                        }
                    >
                        listComponent="ul"
                        listClassName="list-group mb-2"
                        showMoreClassName="mb-0"
                        noun="executor"
                        pluralNoun="executors"
                        querySubject={querySubject}
                        nodeComponent={ExecutorNode}
                        nodeComponentProps={{
                            sourcegraphVersion: data.site.productVersion,
                        }}
                        queryConnection={queryExecutorsCallback}
                        history={history}
                        location={props.location}
                        cursorPaging={true}
                        filters={filters}
                        emptyElement={<NoExecutors />}
                    />
                )}
            </Container>
        </>
    )
}

export interface ExecutorNodeProps {
    node: ExecutorFields
    sourcegraphVersion: string
}

/**
 * Valid build date examples for sourcegraph
 * 169135_2022-08-25_a2b623dce148
 * 169120_2022-08-25_a94c7eb7beca
 *
 * Valid build date example for executor (patch)
 * executor-patch-notest-es-ignite-debug_168065_2022-08-18_e94e18c4ebcc_patch
 */
const buildDateRegex = /^[\w-]+_(\d{4}-\d{2}-\d{2})_\w+/
const developmentVersion = '0.0.0+dev'

export const ExecutorNode: FunctionComponent<React.PropsWithChildren<ExecutorNodeProps>> = ({
    node,
    sourcegraphVersion,
}) => {
    const isOutdated = useMemo(() => {
        const isDevelopment = node.executorVersion === developmentVersion && sourcegraphVersion === developmentVersion

        // We don't need to have this check when in development as the executors will also be
        // in development mode. We also don't need this check for inactive executors.
        if (!isDevelopment && node.active) {
            const semverExecutorVersion = semver.parse(node.executorVersion)
            const semverSourcegraphVersion = semver.parse(sourcegraphVersion)

            if (semverExecutorVersion && semverSourcegraphVersion) {
                // if the sourcegraph version is greater than the executor version, the
                // executor needs to be updated to match the sourcegraph version.
                return semver.gt(semverSourcegraphVersion, semverExecutorVersion)
            }

            // version is not in semver. We need to use the `buildDateRegex` to parse
            // the build date and compare.
            const sourcegraphBuildDateMatch = sourcegraphVersion.match(buildDateRegex)
            const executorBuildDateMatch = node.executorVersion.match(buildDateRegex)

            const isSourcegraphBuildDateValid = sourcegraphBuildDateMatch && sourcegraphBuildDateMatch.length > 1
            const isExecutorBuildDateValid = executorBuildDateMatch && executorBuildDateMatch.length > 1

            if (isSourcegraphBuildDateValid && isExecutorBuildDateValid) {
                const [, sourcegraphBuildDate] = sourcegraphBuildDateMatch
                const [, executorBuildDate] = executorBuildDateMatch

                /**
                 * Syntax: isAfter(date, dateToCompare)
                 *
                 * date	            Date | Number	the date that should be after the other one to return true
                 * dateToCompare	Date | Number	the date to compare with
                 */
                return isAfter(
                    parse(sourcegraphBuildDate, 'yyyy-MM-dd', new Date()),
                    parse(executorBuildDate, 'yyyy-MM-dd', new Date())
                )
            }

            // if all of the above fail, we assume the executor is outdated and something is wrong.
            return true
        }

        return false
    }, [node.active, node.executorVersion, sourcegraphVersion])

    return (
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

            {isOutdated && (
                <Alert variant="warning" className="mt-3">
                    <Text className="m-0">{node.hostname} is outdated.</Text>
                    <Text className="m-0">Please upgrade to a version compatible with your sourcegraph version.</Text>
                </Alert>
            )}
        </li>
    )
}

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
