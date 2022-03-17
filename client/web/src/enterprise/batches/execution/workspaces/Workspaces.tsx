import React, { useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'

import { Scalars } from '../../../../graphql-operations'
import { Header as WorkspacesListHeader } from '../../workspaces-list'
import { useWorkspacesListConnection } from '../backend'

import { WorkspaceFilterRow, WorkspaceFilters } from './WorkspacesFilterRow'
import { WorkspacesList } from './WorkspacesList'

export interface WorkspacesProps {
    batchSpecID: Scalars['ID']
    /** The currently selected workspace node id. Will be highlighted. */
    selectedNode?: Scalars['ID']
    /** The URL path to the execution page this workspaces list is shown on. */
    executionURL: string
}

export const Workspaces: React.FunctionComponent<WorkspacesProps> = ({ batchSpecID, selectedNode, executionURL }) => {
    const [filters, setFilters] = useState<WorkspaceFilters>()
    const workspacesConnection = useWorkspacesListConnection(
        batchSpecID,
        filters?.search ?? null,
        filters?.state ?? null
    )

    return (
        <div className="d-flex flex-column w-100 h-100 pr-3">
            <WorkspacesListHeader>Workspaces</WorkspacesListHeader>
            {workspacesConnection.error && <ErrorAlert error={workspacesConnection.error} className="w-100 mb-0" />}
            <WorkspaceFilterRow onFiltersChange={setFilters} />
            <div className="overflow-auto w-100">
                <WorkspacesList
                    workspacesConnection={workspacesConnection}
                    selectedNode={selectedNode}
                    executionURL={executionURL}
                />
            </div>
        </div>
    )
}
