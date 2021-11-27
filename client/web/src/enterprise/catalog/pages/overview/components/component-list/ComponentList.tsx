import classNames from 'classnames'
import ApplicationCogOutlineIcon from 'mdi-react/ApplicationCogOutlineIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
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
import { CATALOG_COMPONENTS_GQL } from '../../../../core/backend/gql-api/gql/CatalogComponents'
import { CatalogComponentFiltersProps } from '../../../../core/component-filters'

import styles from './ComponentList.module.scss'
import { ComponentListFilters } from './ComponentListFilters'

interface Props extends CatalogComponentFiltersProps {
    /** The currently selected CatalogComponent, if any. */
    selected?: { id: Scalars['ID'] }

    size: 'sm' | 'lg'
    className?: string
}

const FIRST = 20

export const ComponentList: React.FunctionComponent<Props> = ({
    selected,
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
        query: CATALOG_COMPONENTS_GQL,
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
                            selected={selected && node.id === selected.id}
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
    selected: boolean
    size: 'sm' | 'lg'
}> = ({ node, selected, size }) => (
    <li className={classNames('list-group-item d-flex', { active: selected })}>
        <h3 className="h6 font-weight-bold mb-0">
            <Link
                to={`/catalog/${node.id}`}
                className={classNames('d-flex align-items-center', {
                    'text-body': selected,
                    'stretched-link': size === 'sm',
                })}
            >
                <CatalogComponentIcon
                    node={node}
                    className={classNames('icon-inline mr-1', { 'text-muted': !selected })}
                />{' '}
                {node.name}
            </Link>
        </h3>
        <div className="flex-1" />
        {size === 'lg' && node.sourceLocation && <Link to={node.sourceLocation.url}>Source</Link>}
    </li>
)

const CatalogComponentIcon: React.FunctionComponent<{ node: CatalogComponentFields; className?: string }> = ({
    node: { kind },
    className,
}) => <ApplicationCogOutlineIcon className={className} />
