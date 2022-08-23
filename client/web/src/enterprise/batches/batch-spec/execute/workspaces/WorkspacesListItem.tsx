import React from 'react'

import { DiffStat } from '../../../../../components/diff/DiffStat'
import { BatchSpecWorkspaceAndStatusFields, Scalars } from '../../../../../graphql-operations'
import { Descriptor, ListItem } from '../../../workspaces-list'
import { useWorkspaceFromCache } from '../backend'

import { WorkspaceStateIcon } from './WorkspaceStateIcon'

import styles from './WorkspacesListItem.module.scss'

interface WorkspacesListItemProps {
    id: Scalars['ID']
    type: 'VisibleBatchSpecWorkspace' | 'HiddenBatchSpecWorkspace'
    /** Whether or not this item is selected to view the details of. */
    isSelected: boolean
    /** Handler when this item is selected. */
    onSelect: () => void
}

export const WorkspacesListItem: React.FunctionComponent<React.PropsWithChildren<WorkspacesListItemProps>> = ({
    id,
    type,
    isSelected,
    onSelect,
}) => {
    const workspace = useWorkspaceFromCache(id, type)
    if (!workspace) {
        return null
    }

    return <MemoizedWorkspacesListItem workspace={workspace} isSelected={isSelected} onSelect={onSelect} />
}

type MemoizedWorkspacesListItemProps = Pick<WorkspacesListItemProps, 'isSelected' | 'onSelect'> & {
    workspace: BatchSpecWorkspaceAndStatusFields
}

export const MemoizedWorkspacesListItem: React.FunctionComponent<
    React.PropsWithChildren<MemoizedWorkspacesListItemProps>
> = React.memo(function MemoizedWorkspacesListItem({ isSelected, onSelect, workspace }) {
    const statusIndicator = (
        <WorkspaceStateIcon cachedResultFound={workspace.cachedResultFound} state={workspace.state} />
    )

    const diffStat = (
        <>{workspace.diffStat && <DiffStat className="pr-3" {...workspace.diffStat} expandedCounts={true} />}</>
    )

    return (
        <ListItem className={isSelected ? styles.selected : undefined} onClick={onSelect}>
            <Descriptor
                workspace={workspace.__typename === 'HiddenBatchSpecWorkspace' ? undefined : workspace}
                statusIndicator={statusIndicator}
            />
            {diffStat}
        </ListItem>
    )
})
