import classNames from 'classnames'
import React, { useEffect, useMemo } from 'react'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { useConnection } from '../../../../../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../../../../components/FilteredConnection/ui'
import {
    ComponentsForExplorerResult,
    ComponentsForExplorerVariables,
    ComponentListFields,
} from '../../../../../graphql-operations'
import { ComponentFiltersProps } from '../../../core/component-query'

import styles from './CatalogExplorerList.module.scss'
import { ComponentRow, ComponentRowsHeader, CatalogExplorerRowStyleProps } from './ComponentRow'
import { COMPONENT_LIST_FRAGMENT } from './gql'

interface Props extends Pick<ComponentFiltersProps, 'filters'>, CatalogExplorerRowStyleProps {
    queryScope?: string

    /** Called to pass the list of actual tags seen among components to the parent. */
    onTagsChange?: (tags: string[]) => void

    className?: string
}

const FIRST = 20

export const CatalogExplorerList: React.FunctionComponent<Props> = ({
    filters,
    queryScope,
    onTagsChange,
    className,
    itemStartClassName,
    itemEndClassName,
    noBottomBorder,
}) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        ComponentsForExplorerResult,
        ComponentsForExplorerVariables,
        ComponentListFields
    >({
        query: gql`
            query ComponentsForExplorer($query: String, $first: Int, $after: String) {
                components(query: $query, first: $first, after: $after) {
                    nodes {
                        ...ComponentListFields
                    }
                    totalCount
                    pageInfo {
                        hasNextPage
                    }
                    tags {
                        name
                    }
                }
            }
            ${COMPONENT_LIST_FRAGMENT}
        `,
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
            return data.components
        },
    })

    const tags: string[] = useMemo(() => connection?.tags.map(tag => tag.name), [connection])
    useEffect(() => onTagsChange?.(tags), [onTagsChange, tags])

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
                            noun="component"
                            pluralNoun="components"
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
