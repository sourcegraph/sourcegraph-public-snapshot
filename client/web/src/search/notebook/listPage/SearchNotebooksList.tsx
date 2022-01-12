import React, { useCallback } from 'react'
import { useHistory, useLocation } from 'react-router'

import { AuthenticatedUser } from '../../../auth'
import { FilteredConnection, FilteredConnectionFilter } from '../../../components/FilteredConnection'
import {
    ListNotebooksResult,
    ListNotebooksVariables,
    NotebookFields,
    NotebooksOrderBy,
} from '../../../graphql-operations'
import { fetchNotebooks as _fetchNotebooks } from '../backend'

import { NotebookNode, NotebookNodeProps } from './NotebookNode'
import styles from './SearchNotebooksList.module.scss'

interface SearchNotebooksListProps {
    filters: FilteredConnectionFilter[]
    authenticatedUser?: AuthenticatedUser | null
    fetchNotebooks: typeof _fetchNotebooks
}

export const SearchNotebooksList: React.FunctionComponent<SearchNotebooksListProps> = ({
    filters,
    authenticatedUser,
    fetchNotebooks,
}) => {
    const queryConnection = useCallback(
        (args: Partial<ListNotebooksVariables>) => {
            const { orderBy, descending } = args as {
                orderBy: NotebooksOrderBy
                descending: boolean
            }

            return fetchNotebooks({
                first: args.first ?? 10,
                query: args.query ?? undefined,
                after: args.after ?? undefined,
                creatorUserID: authenticatedUser?.id,
                orderBy,
                descending,
            })
        },
        [authenticatedUser, fetchNotebooks]
    )

    const history = useHistory()
    const location = useLocation()

    return (
        <div>
            <FilteredConnection<NotebookFields, Omit<NotebookNodeProps, 'node'>, ListNotebooksResult['notebooks']>
                history={history}
                location={location}
                defaultFirst={10}
                compact={false}
                queryConnection={queryConnection}
                filters={filters}
                hideSearch={false}
                nodeComponent={NotebookNode}
                nodeComponentProps={{
                    location,
                    history,
                }}
                noun="notebook"
                pluralNoun="notebooks"
                noSummaryIfAllNodesVisible={true}
                cursorPaging={true}
                inputClassName={styles.filterInput}
                inputPlaceholder="Search notebooks by title and content"
            />
        </div>
    )
}
