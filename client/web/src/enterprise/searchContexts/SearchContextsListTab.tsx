import React, { useCallback } from 'react'

import classNames from 'classnames'
import { useHistory, useLocation } from 'react-router'

import {
    SearchContextProps,
    ListSearchContextsResult,
    ListSearchContextsVariables,
    SearchContextsOrderBy,
    SearchContextMinimalFields,
} from '@sourcegraph/search'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'

import { AuthenticatedUser } from '../../auth'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionFilterValue,
} from '../../components/FilteredConnection'

import { SearchContextNode, SearchContextNodeProps } from './SearchContextNode'

import styles from './SearchContextsListTab.module.scss'

export interface SearchContextsListTabProps
    extends Pick<
            SearchContextProps,
            'fetchSearchContexts' | 'fetchAutoDefinedSearchContexts' | 'getUserSearchContextNamespaces'
        >,
        PlatformContextProps<'requestGraphQL'> {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser | null
}

export const SearchContextsListTab: React.FunctionComponent<React.PropsWithChildren<SearchContextsListTabProps>> = ({
    isSourcegraphDotCom,
    authenticatedUser,
    getUserSearchContextNamespaces,
    fetchSearchContexts,
    fetchAutoDefinedSearchContexts,
    platformContext,
}) => {
    const queryConnection = useCallback(
        (args: Partial<ListSearchContextsVariables>) => {
            const { namespace, orderBy, descending } = args as {
                namespace: string | undefined
                orderBy: SearchContextsOrderBy
                descending: boolean
            }
            const namespaces = namespace
                ? [namespace === 'global' ? null : namespace]
                : getUserSearchContextNamespaces(authenticatedUser)

            return fetchSearchContexts({
                first: args.first ?? 10,
                query: args.query ?? undefined,
                after: args.after ?? undefined,
                namespaces,
                orderBy,
                descending,
                platformContext,
            })
        },
        [authenticatedUser, fetchSearchContexts, getUserSearchContextNamespaces, platformContext]
    )

    const ownerNamespaceFilterValues: FilteredConnectionFilterValue[] = authenticatedUser
        ? [
              {
                  value: authenticatedUser.id,
                  label: authenticatedUser.username,
                  args: {
                      namespace: authenticatedUser.id,
                  },
              },
              ...authenticatedUser.organizations.nodes.map(org => ({
                  value: org.id,
                  label: org.displayName || org.name,
                  args: {
                      namespace: org.id,
                  },
              })),
          ]
        : []

    const filters: FilteredConnectionFilter[] = [
        {
            label: 'Owner',
            type: 'select',
            id: 'owner',
            tooltip: 'Search context owner',
            values: [
                {
                    value: 'all',
                    label: 'All',
                    args: {},
                },
                {
                    value: 'global-owner',
                    label: 'Global',
                    args: {
                        namespace: 'global',
                    },
                },
                ...ownerNamespaceFilterValues,
            ],
        },
        {
            label: 'Sort by',
            type: 'select',
            id: 'order',
            tooltip: 'Order search contexts',
            values: [
                {
                    value: 'spec-asc',
                    label: 'A-Z',
                    args: {
                        orderBy: SearchContextsOrderBy.SEARCH_CONTEXT_SPEC,
                        descending: false,
                    },
                },
                {
                    value: 'spec-desc',
                    label: 'Z-A',
                    args: {
                        orderBy: SearchContextsOrderBy.SEARCH_CONTEXT_SPEC,
                        descending: true,
                    },
                },
                {
                    value: 'updated-at-asc',
                    label: 'Oldest updates',
                    args: {
                        orderBy: SearchContextsOrderBy.SEARCH_CONTEXT_UPDATED_AT,
                        descending: false,
                    },
                },
                {
                    value: 'updated-at-desc',
                    label: 'Newest updates',
                    args: {
                        orderBy: SearchContextsOrderBy.SEARCH_CONTEXT_UPDATED_AT,
                        descending: true,
                    },
                },
            ],
        },
    ]

    const history = useHistory()
    const location = useLocation()
    return (
        <>
            <FilteredConnection<
                SearchContextMinimalFields,
                Omit<SearchContextNodeProps, 'node'>,
                ListSearchContextsResult['searchContexts']
            >
                history={history}
                location={location}
                defaultFirst={10}
                compact={false}
                queryConnection={queryConnection}
                filters={filters}
                hideSearch={false}
                nodeComponent={SearchContextNode}
                nodeComponentProps={{
                    location,
                    history,
                }}
                noun="search context"
                pluralNoun="search contexts"
                noSummaryIfAllNodesVisible={true}
                cursorPaging={true}
                inputClassName={classNames(styles.filterInput)}
                inputPlaceholder="Filter search contexts..."
                inputAriaLabel="Filter search contexts"
                formClassName={styles.filtersForm}
            />
        </>
    )
}
