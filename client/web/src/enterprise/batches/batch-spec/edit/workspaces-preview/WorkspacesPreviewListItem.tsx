import React, { useCallback, useMemo, useState } from 'react'

import CloseIcon from 'mdi-react/CloseIcon'

import { Button, Icon } from '@sourcegraph/wildcard'

import {
    PreviewHiddenBatchSpecWorkspaceFields,
    PreviewVisibleBatchSpecWorkspaceFields,
} from '../../../../../graphql-operations'
import { CachedIcon, Descriptor, ExcludeIcon, ListItem } from '../../../workspaces-list'

import styles from './WorkspacesPreviewListItem.module.scss'

interface WorkspacesPreviewListItemProps {
    workspace: PreviewVisibleBatchSpecWorkspaceFields | PreviewHiddenBatchSpecWorkspaceFields
    /** Whether or not this item is stale */
    isStale: boolean
    /** Function to automatically update batch spec to exclude this item. */
    exclude: (repo: string, branch: string) => void
    /** Whether or not the item presented should be read-only. */
    isReadOnly?: boolean
}

export const WorkspacesPreviewListItem: React.FunctionComponent<
    React.PropsWithChildren<WorkspacesPreviewListItemProps>
> = ({ workspace, isStale, exclude, isReadOnly = false }) => {
    const [toBeExcluded, setToBeExcluded] = useState(false)

    const handleExclude = useCallback(() => {
        if (workspace.__typename === 'HiddenBatchSpecWorkspace') {
            return
        }
        setToBeExcluded(true)
        exclude(workspace.repository.name, workspace.branch.displayName)
    }, [exclude, workspace])

    const statusIndicator = useMemo(
        () => (toBeExcluded ? <ExcludeIcon /> : workspace.cachedResultFound ? <CachedIcon /> : undefined),
        [toBeExcluded, workspace.cachedResultFound]
    )

    return (
        <ListItem className={!isReadOnly && isStale ? styles.stale : undefined}>
            <Descriptor
                workspace={workspace.__typename === 'HiddenBatchSpecWorkspace' ? undefined : workspace}
                statusIndicator={statusIndicator}
            />
            {isReadOnly || (workspace.__typename !== 'HiddenBatchSpecWorkspace' && toBeExcluded) ? null : (
                <ExcludeButton handleExclude={handleExclude} />
            )}
        </ListItem>
    )
}

const ExcludeButton: React.FunctionComponent<React.PropsWithChildren<{ handleExclude: () => void }>> = ({
    handleExclude,
}) => (
    <Button className="p-0 my-0 mx-2" data-tooltip="Omit this repository from batch spec file" onClick={handleExclude}>
        <Icon as={CloseIcon} />
    </Button>
)
