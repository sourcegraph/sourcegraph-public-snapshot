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
    className?: string
}

const FIRST = 20

export const CatalogExplorerList: React.FunctionComponent<Props> = ({ filters, className }) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        CatalogEntitiesForExplorerResult,
        CatalogEntitiesForExplorerVariables,
        CatalogEntityForExplorerFields
    >({
        query: CATALOG_ENTITIES_FOR_EXPLORER,
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
            return data.catalog.entities
        },
    })

    return (
        <>
            <ConnectionContainer className={className}>
                {error && <ConnectionError errors={[error.message]} />}
                {connection?.nodes && connection?.nodes.length > 0 && (
                    <ConnectionList className={classNames(styles.table)} as="div">
                        <span className="text-muted mt-2 small">Name</span>
                        <span className="text-muted mt-2 small">Owner</span>
                        <span className="text-muted mt-2 small">Lifecycle</span>
                        <span className="text-muted mt-2 small">Description</span>
                        <div className={classNames('border-top', styles.separator)} />
                        {connection?.nodes?.map(node => (
                            <CatalogEntity key={node.id} node={node} />
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
}> = ({ node }) => (
    <>
        <h3 className="h6 font-weight-bold mb-0 d-flex align-items-center">
            <Link to={node.url} className={classNames('d-block text-truncate')}>
                <CatalogEntityIcon entity={node} className={classNames('icon-inline mr-1 flex-shrink-0 text-muted')} />
                {node.name}
            </Link>
            <CatalogEntityStateIndicator entity={node} />
        </h3>
        <EntityOwner owner={node.owner} className="text-nowrap" blankIfNone={true} />
        <span className="text-nowrap">{node.lifecycle.toLowerCase()}</span>
        <div className="text-muted text-truncate">{node.description}</div>
        <div className={classNames('border-top', styles.separator)} />
    </>
)
