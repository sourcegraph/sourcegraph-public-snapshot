import type { ReactElement } from 'react'

import { mdiSourceBranch } from '@mdi/js'
import classNames from 'classnames'

import { Icon, H4, Badge } from '@sourcegraph/wildcard'

import styles from './Descriptor.module.scss'

interface WorkspaceBaseFields {
    branch: {
        displayName: string
    }
    path: string
    repository: {
        name: string
    }
    ignored: boolean
    unsupported: boolean
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
    <div className={styles.container}>
        <div className={styles.status}>{statusIndicator}</div>
        <div className="flex-1">
            <H4 className={styles.name}>{workspace?.repository.name ?? 'Workspace in hidden repository'}</H4>
            {workspace && workspace.path !== '' && workspace.path !== '/' ? (
                <span aria-label={`Workspace path: ${workspace?.path}`} className={styles.path}>
                    {workspace?.path}
                </span>
            ) : null}
            {workspace && (
                <div className={classNames(styles.workspaceDetails, 'text-monospace')}>
                    {workspace.ignored && (
                        <Badge
                            className={styles.badge}
                            variant="secondary"
                            tooltip="This workspace is going to be ignored. A .batchignore file was found in it."
                        >
                            IGNORED
                        </Badge>
                    )}
                    {workspace.unsupported && (
                        <Badge
                            className={styles.badge}
                            variant="secondary"
                            tooltip="This workspace is going to be skipped. It was found on a code-host that is not yet supported by batch changes."
                        >
                            UNSUPPORTED
                        </Badge>
                    )}
                    <Icon aria-hidden={true} className="mr-1" svgPath={mdiSourceBranch} />
                    <small aria-label={`Workspace branch: ${workspace.branch.displayName}`}>
                        {workspace.branch.displayName}
                    </small>
                </div>
            )}
        </div>
    </div>
)
