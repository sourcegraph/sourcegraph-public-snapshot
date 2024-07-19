import { useEffect, useMemo, type FunctionComponent } from 'react'

import { mdiLink } from '@mdi/js'
import classNames from 'classnames'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import {
    Button,
    Container,
    ErrorAlert,
    H2,
    Icon,
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
import { ConnectionForm } from '../components/FilteredConnection/ui'
import { PromptsOrderBy, type PromptFields, type PromptsResult, type PromptsVariables } from '../graphql-operations'
import { LibraryItemStatusBadge, LibraryItemVisibilityBadge } from '../library/itemBadges'
import { useAffiliatedNamespaces } from '../namespaces/useAffiliatedNamespaces'
import { PageRoutes } from '../routes.constants'

import { promptsQuery } from './graphql'
import { PromptNameWithOwner } from './PromptNameWithOwner'
import { urlToEditPrompt } from './util'

import styles from './ListPage.module.scss'

const PromptNode: FunctionComponent<
    TelemetryV2Props & {
        prompt: PromptFields
    }
> = ({ prompt }) => (
    <div className={classNames(styles.row, 'list-group-item align-items-center flex-gap-2')}>
        <div className="flex-1">
            <H2 className="text-base mb-0 font-weight-normal">
                <Link to={prompt.url}>
                    <PromptNameWithOwner prompt={prompt} />
                </Link>
                <LibraryItemVisibilityBadge item={prompt} className="ml-2" />
                <LibraryItemStatusBadge item={prompt} className="ml-2" />
            </H2>
            {prompt.description && <Text className="text-muted text-truncate small mb-0">{prompt.description}</Text>}
        </div>
        <div className="flex-1" />
        <Button to={prompt.url} variant="secondary" as={Link}>
            <Icon aria-label="Permalink" svgPath={mdiLink} />
        </Button>
        {prompt.viewerCanAdminister && (
            <Button to={urlToEditPrompt(prompt)} variant="secondary" as={Link}>
                Edit
            </Button>
        )}
    </div>
)

export function urlToPromptsList(owner: PromptFields['owner']['id']): string {
    return `${PageRoutes.Prompts}?owner=${encodeURIComponent(owner)}`
}

/**
 * List of prompts.
 */
export const ListPage: FunctionComponent<TelemetryV2Props> = ({ telemetryRecorder }) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('prompts.list', 'view')
    }, [telemetryRecorder])

    const { namespaces, loading: namespacesLoading, error: namespacesError } = useAffiliatedNamespaces()
    const filters = useMemo<
        Filter<'drafts' | 'owner' | 'order', Partial<Omit<PromptsVariables, PaginationKeys | 'query'>>>[]
    >(
        () => [
            {
                label: 'Show drafts',
                type: 'select',
                id: 'drafts',
                tooltip: 'Include draft prompts',
                options: [
                    {
                        value: 'true',
                        label: 'Yes',
                        args: {
                            includeDrafts: true,
                        },
                    },
                    {
                        value: 'false',
                        label: 'No',
                        args: {
                            includeDrafts: false,
                        },
                    },
                ],
            },
            {
                label: 'Owner',
                type: 'select',
                id: 'owner',
                tooltip: 'User or organization that owns the prompt',
                options: [
                    {
                        value: 'all',
                        label: 'All',
                        args: {},
                    },
                    ...(namespaces?.map(namespace => ({
                        value: namespace.id,
                        label: namespace.namespaceName,
                        args: {
                            owner: namespace.id,
                        },
                    })) ?? []),
                ],
            },
            {
                label: 'Sort',
                type: 'select',
                id: 'order',
                options: [
                    {
                        value: 'updated-at-desc',
                        label: 'Recently updated',
                        args: {
                            orderBy: PromptsOrderBy.PROMPT_UPDATED_AT,
                        },
                    },
                    {
                        value: 'name-asc',
                        label: 'By name',
                        args: {
                            orderBy: PromptsOrderBy.PROMPT_NAME_WITH_OWNER,
                        },
                    },
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
    } = usePageSwitcherPagination<PromptsResult, Partial<PromptsVariables>, PromptFields, typeof connectionState>({
        query: promptsQuery,
        variables: { ...buildFilterArgs(filters, connectionState), viewerIsAffiliated: true, query: debouncedQuery },
        getConnection: ({ data }) => data?.prompts || undefined,
        state: [connectionState, setConnectionState],
    })

    const error = namespacesError || listError
    const loading = namespacesLoading || listLoading

    return (
        <>
            <Container data-testid="prompts-list-page">
                <ConnectionForm
                    hideSearch={false}
                    showSearchFirst={true}
                    inputClassName="mw-30"
                    inputPlaceholder="Find a prompt..."
                    inputAriaLabel=""
                    inputValue={connectionState.query}
                    onInputChange={event => {
                        setConnectionState(prev => ({ ...prev, query: event.target.value }))
                    }}
                    autoFocus={false}
                    filters={filters}
                    onFilterSelect={(filter, value) => setConnectionState(prev => ({ ...prev, [filter.id]: value }))}
                    filterValues={connectionState}
                    compact={false}
                    formClassName="flex-gap-4 mb-4"
                />
                {loading ? (
                    <LoadingSpinner />
                ) : error ? (
                    <ErrorAlert error={error} className="mb-3" />
                ) : !connection?.nodes || connection.nodes.length === 0 ? (
                    <Text className="text-center text-muted mb-0">No prompts found.</Text>
                ) : (
                    <div className="list-group list-group-flush">
                        {connection.nodes.map(prompt => (
                            <PromptNode key={prompt.id} prompt={prompt} telemetryRecorder={telemetryRecorder} />
                        ))}
                    </div>
                )}
            </Container>
            <PageSwitcher {...paginationProps} className="mt-4" totalCount={connection?.totalCount ?? null} />
        </>
    )
}
