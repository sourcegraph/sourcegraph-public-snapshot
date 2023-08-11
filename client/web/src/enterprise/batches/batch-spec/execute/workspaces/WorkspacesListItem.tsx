import { type FC, useCallback } from 'react'

import { useNavigate } from 'react-router-dom'

import { DiffStat } from '../../../../../components/diff/DiffStat'
import type {
    HiddenBatchSpecWorkspaceListFields,
    Scalars,
    VisibleBatchSpecWorkspaceListFields,
} from '../../../../../graphql-operations'
import { Descriptor, ListItem } from '../../../workspaces-list'

import { WorkspaceStateIcon } from './WorkspaceStateIcon'

import styles from './WorkspacesListItem.module.scss'

export interface WorkspacesListItemProps {
    node: VisibleBatchSpecWorkspaceListFields | HiddenBatchSpecWorkspaceListFields
    /** The currently selected workspace node id. Will be highlighted. */
    selectedNode?: Scalars['ID']
    /** The URL path to the execution page + tab this workspaces list item is shown on. */
    executionURL: string
}

export const WorkspacesListItem: FC<WorkspacesListItemProps> = ({ node: workspace, selectedNode, executionURL }) => {
    const navigate = useNavigate()

    const onSelect = useCallback(
        () => navigate({ pathname: `${executionURL}/execution/workspaces/${workspace.id}` }),
        [navigate, executionURL, workspace.id]
    )

    const statusIndicator = (
        <WorkspaceStateIcon cachedResultFound={workspace.cachedResultFound} state={workspace.state} />
    )

    const diffStat = (
        <>{workspace.diffStat && <DiffStat className="pr-3" {...workspace.diffStat} expandedCounts={true} />}</>
    )

    return (
        <ListItem className={selectedNode === workspace.id ? styles.selected : undefined} onClick={onSelect}>
            <Descriptor
                workspace={workspace.__typename === 'HiddenBatchSpecWorkspace' ? undefined : workspace}
                statusIndicator={statusIndicator}
            />
            {diffStat}
        </ListItem>
    )
}
