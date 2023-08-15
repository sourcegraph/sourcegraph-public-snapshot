import React, { useCallback, useMemo, useState } from 'react'

import { mdiClose } from '@mdi/js'

import { Button, Icon, screenReaderAnnounce, Tooltip } from '@sourcegraph/wildcard'

import type {
    PreviewHiddenBatchSpecWorkspaceFields,
    PreviewVisibleBatchSpecWorkspaceFields,
} from '../../../../../graphql-operations'
import { CachedIcon, Descriptor, ExcludeIcon, ListItem, PartiallyCachedIcon } from '../../../workspaces-list'

import styles from './WorkspacesPreviewListItem.module.scss'

interface WorkspacesPreviewListItemProps {
    workspace: PreviewVisibleBatchSpecWorkspaceFields | PreviewHiddenBatchSpecWorkspaceFields
    /** Whether or not this item is stale */
    isStale: boolean
    /** Function to automatically update batch spec to exclude this item. */
    exclude: (repo: string, branch: string) => void
    /** Whether or not the item presented should be read-only. */
    isReadOnly?: boolean
    /** Whether using cached results is disabled. */
    cacheDisabled?: boolean
}

export const WorkspacesPreviewListItem: React.FunctionComponent<
    React.PropsWithChildren<WorkspacesPreviewListItemProps>
> = ({ workspace, isStale, exclude, cacheDisabled, isReadOnly = false }) => {
    const [toBeExcluded, setToBeExcluded] = useState(false)

    const handleExclude = useCallback(() => {
        if (workspace.__typename === 'HiddenBatchSpecWorkspace') {
            return
        }
        setToBeExcluded(true)
        exclude(workspace.repository.name, workspace.branch.displayName)
        screenReaderAnnounce('Batch spec has been updated to exclude this workspace.')
    }, [exclude, workspace])

    const statusIndicator = useMemo(() => {
        if (toBeExcluded) {
            return <ExcludeIcon />
        }
        if (workspace.cachedResultFound) {
            return <CachedIcon cacheDisabled={cacheDisabled} />
        }
        if (workspace.stepCacheResultCount > 0) {
            return <PartiallyCachedIcon cacheDisabled={cacheDisabled} count={workspace.stepCacheResultCount} />
        }
        return undefined
    }, [cacheDisabled, toBeExcluded, workspace.cachedResultFound, workspace.stepCacheResultCount])

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
    <Tooltip content="Omit this repository from batch spec file">
        <Button aria-label="Omit this repository" className="p-0 my-0 mx-2" onClick={handleExclude}>
            <Icon aria-hidden={true} svgPath={mdiClose} />
        </Button>
    </Tooltip>
)
