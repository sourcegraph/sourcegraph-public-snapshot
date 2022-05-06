import React, { useCallback, useEffect } from 'react'

import { useHistory, useLocation } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { FilteredConnection, FilteredConnectionFilter } from '../../components/FilteredConnection'
import { ListNotebooksResult, ListNotebooksVariables, NotebookFields, NotebooksOrderBy } from '../../graphql-operations'
import { fetchNotebooks as _fetchNotebooks } from '../backend'

import { NotebookNode, NotebookNodeProps } from './NotebookNode'

import styles from './NotebooksList.module.scss'

interface NotebooksListProps extends TelemetryProps {
    logEventName: string
    filters: FilteredConnectionFilter[]
    creatorUserID?: string
    starredByUserID?: string
    namespace?: string
    fetchNotebooks: typeof _fetchNotebooks
}

export const NotebooksList: React.FunctionComponent<React.PropsWithChildren<NotebooksListProps>> = ({
    logEventName,
    filters,
    creatorUserID,
    starredByUserID,
    namespace,
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
                namespace,
                orderBy,
                descending,
            })
        },
        [creatorUserID, starredByUserID, namespace, fetchNotebooks]
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
