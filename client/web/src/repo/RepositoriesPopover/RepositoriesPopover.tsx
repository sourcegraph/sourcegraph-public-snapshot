import React, { useEffect, useState } from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'
import { useConnection } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionForm,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'
import { useDebounce } from '@sourcegraph/wildcard'

import {
    RepositoriesForPopoverResult,
    RepositoriesForPopoverVariables,
    RepositoryPopoverFields,
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

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

export interface RepositoriesPopoverProps {
    /**
     * The current repository (shown as selected in the list), if any.
     */
    currentRepo?: Scalars['ID']
}

export const BATCH_COUNT = 10

/**
 * A popover that displays a searchable list of repositories.
 */
export const RepositoriesPopover: React.FunctionComponent<RepositoriesPopoverProps> = ({ currentRepo }) => {
    const [searchValue, setSearchValue] = useState('')
    const query = useDebounce(searchValue, 200)

    useEffect(() => {
        eventLogger.logViewEvent('RepositoriesPopover')
    }, [])

    const { connection, loading, error, hasNextPage, fetchMore } = useConnection<
        RepositoriesForPopoverResult,
        RepositoriesForPopoverVariables,
        RepositoryPopoverFields
    >({
        query: REPOSITORIES_FOR_POPOVER,
        variables: { first: BATCH_COUNT, after: null, query },
        getConnection: ({ data, errors }) => {
            if (!data || !data.repositories) {
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
        />
    )

    return (
        <div className="repositories-popover connection-popover">
            <ConnectionContainer className="connection-popover__content" compact={true}>
                <ConnectionForm
                    inputValue={searchValue}
                    onInputChange={event => setSearchValue(event.target.value)}
                    inputPlaceholder="Search repositories..."
                    inputClassName="connection-popover__input"
                    autoFocus={true}
                />
                <SummaryContainer>{query && summary}</SummaryContainer>
                {error && <ConnectionError errors={[error.message]} />}
                {connection && (
                    <ConnectionList className="connection-popover__nodes">
                        {connection.nodes.map(node => (
                            <RepositoryNode key={node.id} node={node} currentRepo={currentRepo} />
                        ))}
                    </ConnectionList>
                )}
                {loading && <ConnectionLoading />}
                {!loading && connection && (
                    <SummaryContainer>
                        {!query && summary}
                        {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </div>
    )
}
