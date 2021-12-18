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
    ComponentsForExplorerResult,
    ComponentsForExplorerVariables,
    ComponentForExplorerFields,
} from '../../../../../../graphql-operations'
import { ComponentFiltersProps } from '../../../../core/entity-filters'

import { ComponentRow, ComponentRowsHeader, CatalogExplorerRowStyleProps } from './ComponentRow'
import styles from './CatalogExplorerList.module.scss'
import { COMPONENTS_FOR_EXPLORER } from './gql'

interface Props extends Pick<ComponentFiltersProps, 'filters'>, CatalogExplorerRowStyleProps {
    queryScope?: string
    className?: string
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
        ComponentsForExplorerResult,
        ComponentsForExplorerVariables,
        ComponentForExplorerFields
    >({
        query: COMPONENTS_FOR_EXPLORER,
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
                        <ComponentRowsHeader
                            itemStartClassName={itemStartClassName}
                            itemEndClassName={itemEndClassName}
                        />
                        {connection?.nodes?.map((node, index) => (
                            <ComponentRow
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
                            emptyElement={<p>No results found</p>}
                        />
                        {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </>
    )
}
