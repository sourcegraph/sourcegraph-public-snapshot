import React, { PropsWithChildren, useCallback, useMemo } from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
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

import styles from './SearchContextsList.module.scss'

export interface SearchContextsListProps
    extends Pick<SearchContextProps, 'fetchSearchContexts' | 'getUserSearchContextNamespaces'>,
        PlatformContextProps<'requestGraphQL'> {
    authenticatedUser: AuthenticatedUser | null
    setAlert: (message: string) => void
}

export const SearchContextsList: React.FunctionComponent<SearchContextsListProps> = ({
    authenticatedUser,
    getUserSearchContextNamespaces,
    fetchSearchContexts,
    platformContext,
    setAlert,
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

    const ownerNamespaceFilterValues: FilteredConnectionFilterValue[] = useMemo(
        () =>
            authenticatedUser
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
                : [],
        [authenticatedUser]
    )

    const filters: FilteredConnectionFilter[] = useMemo(
        () => [
            {
                label: 'Sort',
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
        ],
        [ownerNamespaceFilterValues]
    )

    const history = useHistory()
    const location = useLocation()

    return (
        <FilteredConnection<
            SearchContextMinimalFields,
            Omit<SearchContextNodeProps, 'node'>,
            ListSearchContextsResult['searchContexts']
        >
            listComponent="table"
            contentWrapperComponent={SearchContextsTableWrapper}
            headComponent={SearchContextsTableHeader}
            history={history}
            location={location}
            defaultFirst={10}
            compact={false}
            queryConnection={queryConnection}
            filters={filters}
            hideSearch={false}
            showSearchFirst={true}
            nodeComponent={SearchContextNode}
            nodeComponentProps={{
                authenticatedUser,
                setAlert,
            }}
            noun="search context"
            pluralNoun="search contexts"
            cursorPaging={true}
            inputClassName={classNames(styles.filterInput)}
            inputPlaceholder="Find a context"
            inputAriaLabel="Find a context"
            formClassName={styles.filtersForm}
        />
    )
}

const SearchContextsTableWrapper: React.FunctionComponent<PropsWithChildren<{}>> = ({ children }) => (
    <div className={styles.tableWrapper}>{children}</div>
)

const SearchContextsTableHeader: React.FunctionComponent = () => (
    <thead>
        <tr>
            <th>
                <VisuallyHidden>Starred</VisuallyHidden>
            </th>
            <th>Name</th>
            <th>Description</th>
            <th>Contents</th>
            <th>Last updated</th>
            <th>
                <VisuallyHidden>Tags</VisuallyHidden>
            </th>
            <th>
                <VisuallyHidden>Actions</VisuallyHidden>
            </th>
        </tr>
    </thead>
)
