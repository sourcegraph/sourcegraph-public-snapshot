import classNames from 'classnames'
import React from 'react'

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
import { CatalogEntityFiltersProps } from '../../../../core/entity-filters'

import { CatalogEntityRow, CatalogEntityRowsHeader, CatalogExplorerRowStyleProps } from './CatalogEntityRow'
import styles from './CatalogExplorerList.module.scss'
import { CATALOG_ENTITIES_FOR_EXPLORER } from './gql'

interface Props extends Pick<CatalogEntityFiltersProps, 'filters'>, CatalogExplorerRowStyleProps {
    queryScope?: string
    className?: string
}

const FIRST = 20

export const CatalogExplorerRelationList: React.FunctionComponent<Props> = ({
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
                        <CatalogEntityRowsHeader
                            itemStartClassName={itemStartClassName}
                            itemEndClassName={itemEndClassName}
                        />
                        {connection?.nodes?.map((node, index) => (
                            <CatalogEntityRow
                                key={node.id}
                                node={node}
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
