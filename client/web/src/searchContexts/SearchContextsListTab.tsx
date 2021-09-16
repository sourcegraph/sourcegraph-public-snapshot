import classNames from 'classnames'
import React, { useCallback } from 'react'
import { useHistory, useLocation } from 'react-router'
import { catchError } from 'rxjs/operators'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { AuthenticatedUser } from '../auth'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionFilterValue,
} from '../components/FilteredConnection'
import {
    ListSearchContextsResult,
    ListSearchContextsVariables,
    SearchContextFields,
    SearchContextsOrderBy,
} from '../graphql-operations'
import { SearchContextProps } from '../search'

import { SearchContextNode, SearchContextNodeProps } from './SearchContextNode'

export interface SearchContextsListTabProps
    extends Pick<
        SearchContextProps,
        'fetchSearchContexts' | 'fetchAutoDefinedSearchContexts' | 'getUserSearchContextNamespaces'
    > {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser | null
}

export const SearchContextsListTab: React.FunctionComponent<SearchContextsListTabProps> = ({
    isSourcegraphDotCom,
    authenticatedUser,
    getUserSearchContextNamespaces,
    fetchSearchContexts,
    fetchAutoDefinedSearchContexts,
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
            })
        },
        [authenticatedUser, fetchSearchContexts, getUserSearchContextNamespaces]
    )

    const autoDefinedSearchContexts = useObservable(fetchAutoDefinedSearchContexts.pipe(catchError(() => [])))

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
            label: 'Order by',
            type: 'select',
            id: 'order',
            tooltip: 'Order search contexts',
            values: [
                {
                    value: 'spec-asc',
                    label: 'Spec (ascending)',
                    args: {
                        orderBy: SearchContextsOrderBy.SEARCH_CONTEXT_SPEC,
                        descending: false,
                    },
                },
                {
                    value: 'spec-desc',
                    label: 'Spec (descending)',
                    args: {
                        orderBy: SearchContextsOrderBy.SEARCH_CONTEXT_SPEC,
                        descending: true,
                    },
                },
                {
                    value: 'updated-at-asc',
                    label: 'Last update (ascending)',
                    args: {
                        orderBy: SearchContextsOrderBy.SEARCH_CONTEXT_UPDATED_AT,
                        descending: false,
                    },
                },
                {
                    value: 'updated-at-desc',
                    label: 'Last update (descending)',
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
            {isSourcegraphDotCom && (
                <div
                    className={classNames(
                        'search-contexts-list-tab__auto-defined-search-contexts',
                        'mb-4',
                        autoDefinedSearchContexts && autoDefinedSearchContexts.length >= 3
                            ? 'search-contexts-list-tab__auto-defined-search-contexts--repeat-3'
                            : 'search-contexts-list-tab__auto-defined-search-contexts--repeat-2'
                    )}
                >
                    {autoDefinedSearchContexts?.map(context => (
                        <div key={context.spec} className="card p-3">
                            <div>
                                <Link to={`/contexts/${context.spec}`}>
                                    <strong>{context.spec}</strong>
                                </Link>
                                <span
                                    className="badge badge-pill badge-secondary ml-1"
                                    data-tooltip="Automatic contexts are created by Sourcegraph."
                                >
                                    auto
                                </span>
                            </div>
                            <div className="text-muted mt-1">{context.description}</div>
                        </div>
                    ))}
                </div>
            )}

            <FilteredConnection<
                SearchContextFields,
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
                inputClassName="search-contexts-list-tab__filter-input"
                inputPlaceholder="Filter search contexts..."
            />
        </>
    )
}
