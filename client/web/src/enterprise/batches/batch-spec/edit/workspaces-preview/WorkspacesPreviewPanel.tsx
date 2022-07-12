import React from 'react'

import { Panel, useWindowSize } from '@sourcegraph/wildcard'

import { WorkspacesPreview } from './WorkspacesPreview'

import styles from './WorkspacesPreviewPanel.module.scss'

const WORKSPACES_PREVIEW_SIZE = 'batch-changes.ssbc-workspaces-preview-size'

export const WorkspacesPreviewPanel: React.FunctionComponent<React.PropsWithChildren<{ isReadOnly?: boolean }>> = ({
    isReadOnly,
}) => {
    const { width } = useWindowSize()

    return (
        <Panel
            className="d-flex"
            defaultSize={500}
            minSize={405}
            maxSize={0.45 * width}
            position="right"
            storageKey={WORKSPACES_PREVIEW_SIZE}
        >
            <div className={styles.container}>
                <WorkspacesPreview isReadOnly={isReadOnly} />
            </div>
        </Panel>
    )
}
