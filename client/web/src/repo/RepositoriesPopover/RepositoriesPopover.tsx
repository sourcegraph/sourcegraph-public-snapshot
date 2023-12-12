import React, { useEffect, useState } from 'react'

import { createAggregateError } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import type { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useDebounce } from '@sourcegraph/wildcard'

import { useShowMorePagination } from '../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionError,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../components/FilteredConnection/ui'
import type {
    RepositoriesForPopoverResult,
    RepositoriesForPopoverVariables,
    RepositoryPopoverFields,
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import {
    ConnectionPopover,
    ConnectionPopoverContainer,
    ConnectionPopoverForm,
    ConnectionPopoverList,
} from '../RevisionsPopover/components'

import { RepositoryNode } from './RepositoryNode'

export const REPOSITORIES_FOR_POPOVER = gql`
    query RepositoriesForPopover($first: Int, $query: String, $after: String) {
        repositories(first: $first, after: $after, query: $query) {
            nodes {
                ...RepositoryPopoverFields
            }
            pageInfo {
                hasNextPage
                endCursor
            }
        }
    }

    fragment RepositoryPopoverFields on Repository {
        __typename
        id
        name
    }
`

export interface RepositoriesPopoverProps extends TelemetryProps {
    /**
     * The current repository (shown as selected in the list), if any.
     */
    currentRepo?: Scalars['ID']
}

export const BATCH_COUNT = 10

/**
 * A popover that displays a searchable list of repositories.
 */
export const RepositoriesPopover: React.FunctionComponent<React.PropsWithChildren<RepositoriesPopoverProps>> = ({
    currentRepo,
    telemetryService,
    telemetryRecorder,
}) => {
    const [searchValue, setSearchValue] = useState('')
    const query = useDebounce(searchValue, 200)

    useEffect(() => {
        eventLogger.logViewEvent('RepositoriesPopover')
        telemetryService.log('RepositoriesPopover')
        telemetryRecorder.recordEvent('repositoriesPopover', 'viewed')
    }, [telemetryService, telemetryRecorder])

    const { connection, loading, error, hasNextPage, fetchMore } = useShowMorePagination<
        RepositoriesForPopoverResult,
        RepositoriesForPopoverVariables,
        RepositoryPopoverFields
    >({
        query: REPOSITORIES_FOR_POPOVER,
        variables: { first: BATCH_COUNT, after: null, query },
        getConnection: ({ data, errors }) => {
            if (!data?.repositories) {
                throw createAggregateError(errors)
            }
            return data.repositories
        },
        options: {
            fetchPolicy: 'cache-first',
        },
    })

    const summary = connection && (
        <ConnectionSummary
            connection={connection}
            first={BATCH_COUNT}
            noun="repository"
            pluralNoun="repositories"
            hasNextPage={hasNextPage}
            connectionQuery={query}
            noSummaryIfAllNodesVisible={true}
            compact={true}
        />
    )

    return (
        <ConnectionPopover>
            <ConnectionPopoverContainer>
                <ConnectionPopoverForm
                    inputValue={searchValue}
                    onInputChange={event => setSearchValue(event.target.value)}
                    inputPlaceholder="Search repositories..."
                    autoFocus={true}
                    compact={true}
                />
                <SummaryContainer compact={true}>{query && summary}</SummaryContainer>
                {error && <ConnectionError errors={[error.message]} compact={true} />}
                {connection && (
                    <ConnectionPopoverList>
                        {connection.nodes.map(node => (
                            <RepositoryNode key={node.id} node={node} currentRepo={currentRepo} />
                        ))}
                    </ConnectionPopoverList>
                )}
                {loading && <ConnectionLoading compact={true} />}
                {!loading && connection && (
                    <SummaryContainer compact={true}>
                        {!query && summary}
                        {hasNextPage && <ShowMoreButton compact={true} onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionPopoverContainer>
        </ConnectionPopover>
    )
}
