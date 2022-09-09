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
import { queryWorkspacesList as _queryWorkspacesList } from '../backend'

import { WorkspaceFilterRow, WorkspaceFilters } from './WorkspacesFilterRow'
import { WorkspacesListItem, WorkspacesListItemProps } from './WorkspacesListItem'

import styles from './Workspaces.module.scss'

export interface WorkspacesProps {
    batchSpecID: Scalars['ID']
    /** The currently selected workspace node id. Will be highlighted. */
    selectedNode?: Scalars['ID']
    /** The URL path to the execution page this workspaces list is shown on. */
    executionURL: string
    /** For testing only. */
    queryWorkspacesList?: typeof _queryWorkspacesList
}

export const Workspaces: React.FunctionComponent<React.PropsWithChildren<WorkspacesProps>> = ({
    batchSpecID,
    selectedNode,
    executionURL,
    queryWorkspacesList = _queryWorkspacesList,
}) => {
    const history = useHistory()
    const location = useLocation()

    const [filters, setFilters] = useState<WorkspaceFilters>({ state: null, search: null })

    const queryWorkspacesListConnection = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryWorkspacesList({
                batchSpecID,
                first: args.first ?? null,
                after: args.after ?? null,
                search: filters.search ?? null,
                state: filters.state ?? null,
            }).pipe(repeatWhen(notifier => notifier.pipe(delay(2500)))),
        [batchSpecID, filters.search, filters.state, queryWorkspacesList]
    )

    return (
        <div className="d-flex flex-column w-100 h-100 pr-3">
            <WorkspacesListHeader>Workspaces</WorkspacesListHeader>
            <WorkspaceFilterRow onFiltersChange={setFilters} />
            <div className={styles.listContainer}>
                {/* We need to use FilteredConnection over the new composable API here because we
                still have to use observables for this connection query. See query docstring for
                more details. */}
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
                    defaultFirst={20}
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
