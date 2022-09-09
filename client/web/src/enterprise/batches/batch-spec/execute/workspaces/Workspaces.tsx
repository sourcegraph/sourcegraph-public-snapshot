import React, { useCallback, useState } from 'react'

import { useHistory, useLocation } from 'react-router'
import { delay, repeatWhen } from 'rxjs/operators'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../../../components/FilteredConnection'
import {
    Scalars,
    VisibleBatchSpecWorkspaceListFields,
    HiddenBatchSpecWorkspaceListFields,
} from '../../../../../graphql-operations'
import { Header as WorkspacesListHeader } from '../../../workspaces-list'
import { queryWorkspacesList } from '../backend'

import { WorkspaceFilterRow, WorkspaceFilters } from './WorkspacesFilterRow'
import { WorkspacesListItem, WorkspacesListItemProps } from './WorkspacesListItem'

import styles from './Workspaces.module.scss'

export interface WorkspacesProps {
    batchSpecID: Scalars['ID']
    /** The currently selected workspace node id. Will be highlighted. */
    selectedNode?: Scalars['ID']
    /** The URL path to the execution page this workspaces list is shown on. */
    executionURL: string
}

export const Workspaces: React.FunctionComponent<React.PropsWithChildren<WorkspacesProps>> = ({
    batchSpecID,
    selectedNode,
    executionURL,
}) => {
    const history = useHistory()
    const location = useLocation()

    const [filters, setFilters] = useState<WorkspaceFilters>()

    const queryWorkspacesListConnection = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryWorkspacesList({
                batchSpecID,
                first: args.first ?? null,
                after: args.after ?? null,
                search: filters?.search ?? null,
                state: filters?.state ?? null,
            }).pipe(repeatWhen(notifier => notifier.pipe(delay(5000)))),
        [batchSpecID, filters]
    )

    return (
        <div className="d-flex flex-column w-100 h-100 pr-3">
            <WorkspacesListHeader>Workspaces</WorkspacesListHeader>
            <WorkspaceFilterRow onFiltersChange={setFilters} />
            <div className={styles.listContainer}>
                <FilteredConnection<
                    VisibleBatchSpecWorkspaceListFields | HiddenBatchSpecWorkspaceListFields,
                    Omit<WorkspacesListItemProps, 'node'>
                >
                    nodeComponent={WorkspacesListItem}
                    nodeComponentProps={{
                        selectedNode,
                        executionURL,
                    }}
                    queryConnection={queryWorkspacesListConnection}
                    hideSearch={true}
                    history={history}
                    location={location}
                    defaultFirst={5}
                    noun="workspace"
                    pluralNoun="workspaces"
                    useURLQuery={true}
                    noSummaryIfAllNodesVisible={true}
                    withCenteredSummary={true}
                />
            </div>
        </div>
    )
}
