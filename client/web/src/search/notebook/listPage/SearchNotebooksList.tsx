import React, { useCallback, useEffect } from 'react'
import { useHistory, useLocation } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

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

interface SearchNotebooksListProps extends TelemetryProps {
    logEventName: string
    filters: FilteredConnectionFilter[]
    creatorUserID?: string
    starredByUserID?: string
    fetchNotebooks: typeof _fetchNotebooks
}

export const SearchNotebooksList: React.FunctionComponent<SearchNotebooksListProps> = ({
    logEventName,
    filters,
    creatorUserID,
    starredByUserID,
    fetchNotebooks,
    telemetryService,
}) => {
    useEffect(() => telemetryService.logViewEvent(`SearchNotebooksList${logEventName}`), [
        logEventName,
        telemetryService,
    ])

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
                creatorUserID,
                starredByUserID,
                orderBy,
                descending,
            })
        },
        [creatorUserID, starredByUserID, fetchNotebooks]
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
