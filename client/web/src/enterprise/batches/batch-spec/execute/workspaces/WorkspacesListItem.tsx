import React, { useCallback } from 'react'

import { useHistory } from 'react-router'

import { DiffStat } from '../../../../../components/diff/DiffStat'
import {
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

export const WorkspacesListItem: React.FunctionComponent<React.PropsWithChildren<WorkspacesListItemProps>> = ({
    node: workspace,
    selectedNode,
    executionURL,
}) => {
    const history = useHistory()
    const onSelect = useCallback(
        () => history.push({ ...history.location, pathname: `${executionURL}/execution/workspaces/${workspace.id}` }),
        [history, executionURL, workspace.id]
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
