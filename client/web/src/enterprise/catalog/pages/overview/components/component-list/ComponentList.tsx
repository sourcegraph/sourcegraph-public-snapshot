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
    CatalogComponentsResult,
    CatalogComponentsVariables,
    CatalogComponentFields,
} from '../../../../../../graphql-operations'
import { CatalogComponentIcon } from '../../../../components/CatalogComponentIcon'
import { CatalogComponentFiltersProps } from '../../../../core/component-filters'

import styles from './ComponentList.module.scss'
import { ComponentListFilters } from './ComponentListFilters'
import { CATALOG_COMPONENTS } from './gql'

interface Props extends CatalogComponentFiltersProps {
    /** The name of the currently selected CatalogComponent, if any. */
    selectedComponentName?: string

    size: 'sm' | 'lg'
    className?: string
}

const FIRST = 20

export const ComponentList: React.FunctionComponent<Props> = ({
    selectedComponentName,
    filters,
    onFiltersChange,
    size,
    className,
}) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        CatalogComponentsResult,
        CatalogComponentsVariables,
        CatalogComponentFields
    >({
        query: CATALOG_COMPONENTS,
        variables: {
            query: filters.query || '',
            first: FIRST,
            after: null,
        },
        options: {
            useURL: true,
            fetchPolicy: 'cache-and-network',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            return data.catalog.components
        },
    })

    return (
        <>
            <ComponentListFilters
                filters={filters}
                onFiltersChange={onFiltersChange}
                size={size}
                className="p-2 border-bottom"
            />
            <ConnectionContainer className={className}>
                {error && <ConnectionError errors={[error.message]} />}
                <ConnectionList className={classNames('list-group list-group-flush', styles.list)}>
                    {connection?.nodes?.map(node => (
                        <CatalogComponent
                            key={node.id}
                            node={node}
                            selected={Boolean(selectedComponentName && node.name === selectedComponentName)}
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
                            noun="component"
                            pluralNoun="components"
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

const CatalogComponent: React.FunctionComponent<{
    node: CatalogComponentFields
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
                <CatalogComponentIcon
                    catalogComponent={node}
                    className={classNames('icon-inline mr-1 flex-shrink-0', { 'text-muted': !selected })}
                />
                {node.name}
            </Link>
        </h3>
        <div className="flex-1" />
        {size === 'lg' && <Link to="TODO">Source</Link>}
    </li>
)
