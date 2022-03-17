import React, { ReactElement } from 'react'

import { Button } from '@sourcegraph/wildcard'

import { Branch } from '../Branch'

import styles from './Descriptor.module.scss'

interface WorkspaceBaseFields {
    branch: {
        displayName: string
    }
    path: string
    repository: {
        name: string
    }
}

interface DescriptorProps<Workspace extends WorkspaceBaseFields> {
    workspace: Workspace
    /** An optional status indicator to display in line with the workspace details. */
    statusIndicator?: JSX.Element
    /** An optional handler for when the workspace name is clicked. */
    onClick?: () => void
}

export const Descriptor = <Workspace extends WorkspaceBaseFields>({
    statusIndicator,
    workspace,
    onClick,
}: DescriptorProps<Workspace>): ReactElement => {
    const repositoryName =
        typeof onClick === 'undefined' ? (
            <h4>{workspace.repository.name}</h4>
        ) : (
            <Button variant="link" onClick={onClick}>
                <h4>{workspace.repository.name}</h4>
            </Button>
        )

    return (
        <div className={styles.descriptor}>
            <div className={styles.status}>{statusIndicator}</div>
            <div className={styles.name}>{repositoryName}</div>
            <div className={styles.pathAndBranch}>
                {workspace.path !== '' && workspace.path !== '/' ? (
                    <span className={styles.path}>{workspace.path}</span>
                ) : null}
                <Branch name={workspace.branch.displayName} />
            </div>
        </div>
    )
}
