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
    ComponentRelationsForExplorerResult,
    ComponentRelationsForExplorerVariables,
    ComponentRelationFields,
} from '../../../../../../graphql-operations'
import { ComponentFiltersProps } from '../../../../core/component-query'

import styles from './CatalogExplorerRelationList.module.scss'
import { ComponentRelationRow, ComponentRelationRowsHeader, CatalogExplorerRowStyleProps } from './ComponentRow'
import { COMPONENT_RELATIONS_FOR_EXPLORER } from './gqlForRelation'

interface Props extends Pick<ComponentFiltersProps, 'filters'>, CatalogExplorerRowStyleProps {
    component: Scalars['ID']
    useURLForConnectionParams?: boolean
    queryScope?: string
    className?: string
}

const FIRST = 20

export const CatalogExplorerRelationList: React.FunctionComponent<Props> = ({
    filters,
    component,
    useURLForConnectionParams,
    queryScope,
    className,
    itemStartClassName,
    itemEndClassName,
    noBottomBorder,
}) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        ComponentRelationsForExplorerResult,
        ComponentRelationsForExplorerVariables,
        ComponentRelationFields
    >({
        query: COMPONENT_RELATIONS_FOR_EXPLORER,
        variables: {
            component,
            query: `${queryScope || ''} ${filters.query || ''}`,
            first: FIRST,
            after: null,
        },
        options: {
            useURL: useURLForConnectionParams,
            fetchPolicy: 'cache-and-network',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            if (!data.node || !('relatedEntities' in data.node)) {
                throw new Error('not a component')
            }
            // TODO(sqs): hack because this connection (correctly per the GraphQL connection spec)
            // returns a field `edges` not `nodes`
            return {
                ...data.node.relatedEntities,
                nodes: [...data.node.relatedEntities.edges]
                    .sort((a, b) => b.type.localeCompare(a.type))
                    .filter(edge => edge.node.id !== component),
            }
        },
    })

    return (
        <>
            <ConnectionContainer className={className}>
                {error && <ConnectionError errors={[error.message]} />}
                {connection?.nodes && connection?.nodes.length > 0 && (
                    <ConnectionList className={classNames(styles.table)} as="div">
                        <ComponentRelationRowsHeader
                            itemStartClassName={itemStartClassName}
                            itemEndClassName={itemEndClassName}
                        />
                        {connection?.nodes?.map((edge, index) => (
                            <ComponentRelationRow
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
