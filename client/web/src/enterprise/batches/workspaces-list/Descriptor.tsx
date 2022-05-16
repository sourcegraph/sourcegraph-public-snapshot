import { ReactElement } from 'react'

import SourceBranchIcon from 'mdi-react/SourceBranchIcon'

import { Icon, Typography } from '@sourcegraph/wildcard'

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
    workspace?: Workspace
    /** An optional status indicator to display in line with the workspace details. */
    statusIndicator?: JSX.Element
}

export const Descriptor = <Workspace extends WorkspaceBaseFields>({
    statusIndicator,
    workspace,
}: DescriptorProps<Workspace>): ReactElement => (
    <div className="d-flex flex-1 align-items-center pt-3 pb-3 pl-2 pr-2">
        <div className={styles.status}>{statusIndicator}</div>
        <div className="flex-1">
            <Typography.H4 className={styles.name}>
                {workspace?.repository.name ?? 'Workspace in hidden repository'}
            </Typography.H4>
            {workspace && workspace.path !== '' && workspace.path !== '/' ? (
                <span className={styles.path}>{workspace?.path}</span>
            ) : null}
            {workspace && (
                <div className="d-flex align-items-center text-muted text-monospace pt-1">
                    <Icon className="mr-1" as={SourceBranchIcon} />
                    <small>{workspace.branch.displayName}</small>
                </div>
            )}
        </div>
    </div>
)
