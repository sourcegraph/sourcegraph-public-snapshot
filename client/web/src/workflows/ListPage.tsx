import { useEffect, useMemo, type FunctionComponent } from 'react'

import classNames from 'classnames'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import {
    Button,
    Container,
    ErrorAlert,
    H2,
    Link,
    LoadingSpinner,
    PageSwitcher,
    Text,
    useDebounce,
} from '@sourcegraph/wildcard'

import { buildFilterArgs, type Filter } from '../components/FilteredConnection'
import { useUrlSearchParamsForConnectionState } from '../components/FilteredConnection/hooks/connectionState'
import {
    usePageSwitcherPagination,
    type PaginationKeys,
} from '../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { ConnectionContainer, ConnectionForm } from '../components/FilteredConnection/ui'
import {
    WorkflowsOrderBy,
    type WorkflowFields,
    type WorkflowsResult,
    type WorkflowsVariables,
} from '../graphql-operations'
import { useAffiliatedNamespaces } from '../namespaces/useAffiliatedNamespaces'
import { PageRoutes } from '../routes.constants'

import { workflowsQuery } from './graphql'
import { WorkflowNameWithOwner } from './WorkflowNameWithOwner'

import styles from './ListPage.module.scss'

const WorkflowNode: FunctionComponent<
    TelemetryV2Props & {
        workflow: WorkflowFields
    }
> = ({ workflow }) => (
    <div className={classNames(styles.row, 'list-group-item test-workflow-list-page-row')}>
        <div className="flex-1">
            <H2 className="text-base mb-0 font-weight-normal">
                <WorkflowNameWithOwner workflow={workflow} />
            </H2>
        </div>
        <div className="flex-0">
            <Button to={workflow.id} variant="secondary" as={Link}>
                Edit
            </Button>
        </div>
    </div>
)

export function urlToWorkflowsList(owner: WorkflowFields['owner']['id']): string {
    return `${PageRoutes.Workflows}?owner=${encodeURIComponent(owner)}`
}

/**
 * List of workflows.
 */
export const ListPage: FunctionComponent<TelemetryV2Props> = ({ telemetryRecorder }) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('workflows.list', 'view')
    }, [telemetryRecorder])

    const { namespaces, loading: namespacesLoading, error: namespacesError } = useAffiliatedNamespaces()
    const filters = useMemo<
        Filter<
            Exclude<keyof WorkflowsVariables, PaginationKeys | 'query'>,
            Partial<Omit<WorkflowsVariables, PaginationKeys | 'query'>>
        >[]
    >(
        () => [
            {
                label: 'Sort',
                type: 'select',
                id: 'orderBy',
                options: [
                    {
                        value: 'updated-at-desc',
                        label: 'Recently updated',
                        args: {
                            orderBy: WorkflowsOrderBy.WORKFLOW_UPDATED_AT,
                        },
                    },
                    {
                        value: 'description-asc',
                        label: 'By description',
                        args: {
                            orderBy: WorkflowsOrderBy.WORKFLOW_NAME_WITH_OWNER,
                        },
                    },
                ],
            },
            {
                label: 'Owner',
                type: 'select',
                id: 'owner',
                tooltip: 'User or organization that owns the workflow',
                options: [
                    {
                        value: 'all',
                        label: 'All',
                        args: {},
                    },
                    ...(namespaces?.map(namespace => ({
                        value: namespace.id,
                        label: (namespace.__typename === 'Org' && namespace.displayName) || namespace.namespaceName,
                        args: {
                            owner: namespace.id,
                        },
                    })) ?? []),
                ],
            },
        ],
        [namespaces]
    )

    const [connectionState, setConnectionState] = useUrlSearchParamsForConnectionState(filters)
    const debouncedQuery = useDebounce(connectionState.query, 300)
    const {
        connection,
        loading: listLoading,
        error: listError,
        ...paginationProps
    } = usePageSwitcherPagination<WorkflowsResult, Partial<WorkflowsVariables>, WorkflowFields, typeof connectionState>(
        {
            query: workflowsQuery,
            variables: { ...buildFilterArgs(filters, connectionState), query: debouncedQuery },
            getConnection: ({ data }) => data?.workflows || undefined,
            state: [connectionState, setConnectionState],
        }
    )

    const error = namespacesError || listError
    const loading = namespacesLoading || listLoading

    return (
        <>
            <Container data-testid="workflows-list-page">
                <ConnectionContainer>
                    <ConnectionForm
                        hideSearch={false}
                        showSearchFirst={true}
                        inputClassName="mw-30"
                        inputPlaceholder="Find a workflow..."
                        inputAriaLabel=""
                        inputValue={connectionState.query}
                        onInputChange={event => {
                            setConnectionState(prev => ({ ...prev, query: event.target.value }))
                        }}
                        autoFocus={false}
                        filters={filters}
                        onFilterSelect={(filter, value) =>
                            setConnectionState(prev => ({ ...prev, [filter.id]: value }))
                        }
                        filterValues={connectionState}
                        compact={false}
                        formClassName="flex-gap-4 mb-4"
                    />
                    {loading ? (
                        <LoadingSpinner />
                    ) : error ? (
                        <ErrorAlert error={error} className="mb-3" />
                    ) : !connection?.nodes || connection.nodes.length === 0 ? (
                        <Text className="text-center text-muted mb-0">No workflows found.</Text>
                    ) : (
                        <div className="list-group list-group-flush">
                            {connection.nodes.map(workflow => (
                                <WorkflowNode
                                    key={workflow.id}
                                    workflow={workflow}
                                    telemetryRecorder={telemetryRecorder}
                                />
                            ))}
                        </div>
                    )}
                </ConnectionContainer>
            </Container>
            <PageSwitcher {...paginationProps} className="mt-4" totalCount={connection?.totalCount ?? null} />
        </>
    )
}
