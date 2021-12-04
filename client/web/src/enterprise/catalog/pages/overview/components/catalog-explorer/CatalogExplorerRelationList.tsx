import classNames from 'classnames'
import React from 'react'

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
    CatalogEntityRelationsForExplorerResult,
    CatalogEntityRelationsForExplorerVariables,
    CatalogEntityRelationFields,
} from '../../../../../../graphql-operations'
import { CatalogEntityFiltersProps } from '../../../../core/entity-filters'

import {
    CatalogEntityRelationRow,
    CatalogEntityRelationRowsHeader,
    CatalogExplorerRowStyleProps,
} from './CatalogEntityRow'
import styles from './CatalogExplorerList.module.scss'
import { CATALOG_ENTITY_RELATIONS_FOR_EXPLORER } from './gqlForRelation'

interface Props extends Pick<CatalogEntityFiltersProps, 'filters'>, CatalogExplorerRowStyleProps {
    entity: Scalars['ID']
    queryScope?: string
    className?: string
}

const FIRST = 20

export const CatalogExplorerRelationList: React.FunctionComponent<Props> = ({
    filters,
    entity,
    queryScope,
    className,
    itemStartClassName,
    itemEndClassName,
    noBottomBorder,
}) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        CatalogEntityRelationsForExplorerResult,
        CatalogEntityRelationsForExplorerVariables,
        CatalogEntityRelationFields
    >({
        query: CATALOG_ENTITY_RELATIONS_FOR_EXPLORER,
        variables: {
            entity,
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
            if (!data.node || !('relatedEntities' in data.node)) {
                throw new Error('not a catalog entity')
            }
            // TODO(sqs): hack because this connection (correctly per the GraphQL connection spec)
            // returns a field `edges` not `nodes`
            return {
                ...data.node.relatedEntities,
                nodes: [...data.node.relatedEntities.edges]
                    .sort((a, b) => b.type.localeCompare(a.type))
                    .filter(edge => edge.node.id !== entity),
            }
        },
    })

    return (
        <>
            <ConnectionContainer className={className}>
                {error && <ConnectionError errors={[error.message]} />}
                {connection?.nodes && connection?.nodes.length > 0 && (
                    <ConnectionList className={classNames(styles.table)} as="div">
                        <CatalogEntityRelationRowsHeader
                            itemStartClassName={itemStartClassName}
                            itemEndClassName={itemEndClassName}
                        />
                        {connection?.nodes?.map((edge, index) => (
                            <CatalogEntityRelationRow
                                key={`${edge.node.id}:${edge.type}`}
                                edge={edge}
                                itemStartClassName={itemStartClassName}
                                itemEndClassName={itemEndClassName}
                                noBottomBorder={index === connection?.nodes?.length - 1 && noBottomBorder}
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
