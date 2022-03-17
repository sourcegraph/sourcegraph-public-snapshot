import React from 'react'

import { DiffStat } from '../../../../components/diff/DiffStat'
import { BatchSpecWorkspaceListFields } from '../../../../graphql-operations'
import { Descriptor, ListItem } from '../../workspaces-list'
import { WorkspaceStateIcon } from '../WorkspaceStateIcon'

import styles from './WorkspacesListItem.module.scss'

interface WorkspacesListItemProps {
    workspace: BatchSpecWorkspaceListFields
    /** Whether or not this item is selected to view the details of. */
    isSelected: boolean
    /** Handler when this item is selected. */
    onSelect: () => void
}

export const WorkspacesListItem: React.FunctionComponent<WorkspacesListItemProps> = ({
    workspace,
    isSelected,
    onSelect,
}) => (
    <ListItem className={isSelected ? styles.selected : undefined}>
        <Descriptor
            workspace={workspace}
            statusIndicator={
                <WorkspaceStateIcon cachedResultFound={workspace.cachedResultFound} state={workspace.state} />
            }
            onClick={onSelect}
        />
        {workspace.diffStat && <DiffStat className="pr-3" {...workspace.diffStat} expandedCounts={true} />}
    </ListItem>
)
