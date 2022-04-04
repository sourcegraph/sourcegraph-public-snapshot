import React, { useCallback, useMemo, useState } from 'react'

import CloseIcon from 'mdi-react/CloseIcon'

import { Button, Icon } from '@sourcegraph/wildcard'

import {
    PreviewHiddenBatchSpecWorkspaceFields,
    PreviewVisibleBatchSpecWorkspaceFields,
} from '../../../../graphql-operations'
import { CachedIcon, Descriptor, ExcludeIcon, ListItem } from '../../workspaces-list'

import styles from './WorkspacesPreviewListItem.module.scss'

interface WorkspacesPreviewListItemProps {
    workspace: PreviewVisibleBatchSpecWorkspaceFields | PreviewHiddenBatchSpecWorkspaceFields
    /** Whether or not this item is stale */
    isStale: boolean
    /** Function to automatically update batch spec to exclude this item. */
    exclude: (repo: string, branch: string) => void
}

export const WorkspacesPreviewListItem: React.FunctionComponent<WorkspacesPreviewListItemProps> = ({
    workspace,
    isStale,
    exclude,
}) => {
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

    if (workspace.__typename === 'HiddenBatchSpecWorkspace') {
        return (
            <li
                className={classNames(
                    'd-flex align-items-center px-2 py-3 w-100',
                    variant === 'light' ? styles.light : styles.dark
                )}
                key={workspace.id}
            >
                <div className={classNames(styles.statusContainer, 'mr-2')}>
                    <StatusIcon
                        status={toBeExcluded ? 'to-exclude' : workspace.cachedResultFound ? 'cached' : 'none'}
                    />
                </div>
                <div className="flex-1">
                    <h4 className={classNames(styles.overflow, (toBeExcluded || isStale) && styles.stale)}>
                        Workspace in hidden repository
                    </h4>
                </div>
            </li>
        )
    }

    return (
        <ListItem className={isStale ? styles.stale : undefined}>
            <Descriptor workspace={workspace} statusIndicator={statusIndicator} />
            {toBeExcluded ? null : <ExcludeButton handleExclude={handleExclude} />}
        </ListItem>
    )
}

const ExcludeButton: React.FunctionComponent<{ handleExclude: () => void }> = ({ handleExclude }) => (
    <Button className="p-0 my-0 mx-2" data-tooltip="Omit this repository from batch spec file" onClick={handleExclude}>
        <Icon as={CloseIcon} />
    </Button>
)
