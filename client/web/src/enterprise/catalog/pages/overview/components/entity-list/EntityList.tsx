import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'

import { useConnection } from '../../../../../../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../../../../../components/FilteredConnection/ui'
import {
    CatalogEntitiesResult,
    CatalogEntitiesVariables,
    CatalogEntityFields,
} from '../../../../../../graphql-operations'
import { CatalogEntityIcon } from '../../../../components/CatalogEntityIcon'
import { CatalogEntityFiltersProps } from '../../../../core/entity-filters'

import { EntityListFilters } from './EntityListFilters'
import { CATALOG_ENTITIES } from './gql'

interface Props extends CatalogEntityFiltersProps {
    /** The name of the currently selected catalog entity, if any. */
    selectedEntityName?: string

    size: 'sm' | 'lg'
    className?: string
}

const FIRST = 20

export const EntityList: React.FunctionComponent<Props> = ({
    selectedEntityName,
    filters,
    onFiltersChange,
    size,
    className,
}) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        CatalogEntitiesResult,
        CatalogEntitiesVariables,
        CatalogEntityFields
    >({
        query: CATALOG_ENTITIES,
        variables: {
            query: filters.query || '',
            first: FIRST,
            after: null,
        },
        options: {
            useURL: size === 'lg',
            fetchPolicy: 'cache-and-network',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            return data.catalog.entities
        },
    })

    return (
        <>
            <EntityListFilters
                filters={filters}
                onFiltersChange={onFiltersChange}
                size={size}
                className="p-2 border-bottom"
            />
            <ConnectionContainer className={className}>
                {error && <ConnectionError errors={[error.message]} />}
                <ConnectionList className={classNames('list-group list-group-flush')}>
                    {connection?.nodes?.map(node => (
                        <CatalogEntity
                            key={node.id}
                            node={node}
                            selected={Boolean(selectedEntityName && node.name === selectedEntityName)}
                            size={size}
                        />
                    ))}
                </ConnectionList>
                {loading && <ConnectionLoading />}
                {connection && (
                    <SummaryContainer centered={true}>
                        <ConnectionSummary
                            noSummaryIfAllNodesVisible={true}
                            first={FIRST}
                            connection={connection}
                            noun="entity"
                            pluralNoun="entities"
                            hasNextPage={hasNextPage}
                            emptyElement={<p>No components found</p>}
                        />
                        {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </>
    )
}

const CatalogEntity: React.FunctionComponent<{
    node: CatalogEntityFields
    selected?: boolean
    size: 'sm' | 'lg'
}> = ({ node, selected, size }) => (
    <li className={classNames('list-group-item d-flex', { active: selected })}>
        <h3 className="h6 font-weight-bold mb-0 overflow-hidden">
            <Link
                to={node.url}
                className={classNames('d-block text-truncate', {
                    'text-body': selected,
                    'stretched-link': size === 'sm',
                })}
            >
                <CatalogEntityIcon
                    entity={node}
                    className={classNames('icon-inline mr-1 flex-shrink-0', { 'text-muted': !selected })}
                />
                {node.name}
            </Link>
        </h3>
        <div className="flex-1" />
    </li>
)
