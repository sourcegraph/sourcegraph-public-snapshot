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
    CatalogEntitiesForExplorerResult,
    CatalogEntitiesForExplorerVariables,
    CatalogEntityForExplorerFields,
} from '../../../../../../graphql-operations'
import { CatalogEntityIcon } from '../../../../components/CatalogEntityIcon'
import { EntityOwner } from '../../../../components/entity-owner/EntityOwner'
import { CatalogEntityFiltersProps } from '../../../../core/entity-filters'
import { CatalogEntityStateIndicator } from '../entity-state-indicator/EntityStateIndicator'

import styles from './CatalogExplorerList.module.scss'
import { CATALOG_ENTITIES_FOR_EXPLORER } from './gql'

interface Props extends Pick<CatalogEntityFiltersProps, 'filters'> {
    queryScope?: string
    className?: string
    noBottomBorder?: boolean
    itemStartClassName?: string
    itemEndClassName?: string
}

const FIRST = 20

export const CatalogExplorerList: React.FunctionComponent<Props> = ({
    filters,
    queryScope,
    className,
    itemStartClassName,
    itemEndClassName,
    noBottomBorder,
}) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        CatalogEntitiesForExplorerResult,
        CatalogEntitiesForExplorerVariables,
        CatalogEntityForExplorerFields
    >({
        query: CATALOG_ENTITIES_FOR_EXPLORER,
        variables: {
            query: `${queryScope || ''} ${filters.query || ''}`,
            first: FIRST,
            after: null,
        },
        options: {
            useURL: true,
            fetchPolicy: 'cache-and-network',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            return data.catalog.entities
        },
    })

    return (
        <>
            <ConnectionContainer className={className}>
                {error && <ConnectionError errors={[error.message]} />}
                {connection?.nodes && connection?.nodes.length > 0 && (
                    <ConnectionList className={classNames(styles.table)} as="div">
                        <div className={classNames('text-muted mt-2 small', itemStartClassName)}>Name</div>
                        <div className="text-muted mt-2 small">Owner</div>
                        <div className="text-muted mt-2 small">Lifecycle</div>
                        <div className={classNames('text-muted mt-2 small', itemEndClassName)}>Description</div>
                        <div className={classNames('border-top', styles.separator)} />
                        {connection?.nodes?.map((node, index) => (
                            <CatalogEntity
                                key={node.id}
                                node={node}
                                itemStartClassName={itemStartClassName}
                                itemEndClassName={itemEndClassName}
                                noBorder={index === connection?.nodes?.length - 1 && noBottomBorder}
                            />
                        ))}
                    </ConnectionList>
                )}
                {loading && <ConnectionLoading className="my-2" />}
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
    node: CatalogEntityForExplorerFields
    itemStartClassName?: string
    itemEndClassName?: string
    noBorder?: boolean
}> = ({ node, itemStartClassName, itemEndClassName, noBorder }) => (
    <>
        <h3 className={classNames('h6 font-weight-bold mb-0 d-flex align-items-center', itemStartClassName)}>
            <Link to={node.url} className={classNames('d-block text-truncate')}>
                <CatalogEntityIcon entity={node} className={classNames('icon-inline mr-1 flex-shrink-0 text-muted')} />
                {node.name}
            </Link>
            <CatalogEntityStateIndicator entity={node} className="ml-1" />
        </h3>
        <EntityOwner owner={node.owner} className="text-nowrap" blankIfNone={true} />
        <span className="text-nowrap">{node.lifecycle.toLowerCase()}</span>
        <div className={classNames('text-muted text-truncate', itemEndClassName)}>{node.description}</div>
        <div className={classNames({ 'border-top': !noBorder }, styles.separator)} />
    </>
)
