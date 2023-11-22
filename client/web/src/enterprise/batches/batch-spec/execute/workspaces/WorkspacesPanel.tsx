import React from 'react'

import { Panel, useWindowSize, VIEWPORT_LG } from '@sourcegraph/wildcard'

import { Workspaces, type WorkspacesProps } from './Workspaces'

import styles from './WorkspacesPanel.module.scss'

const WORKSPACES_LIST_SIZE = 'batch-changes.ssbc-workspaces-list-size'

export const WorkspacesPanel: React.FunctionComponent<React.PropsWithChildren<WorkspacesProps>> = props => {
    const { width } = useWindowSize()

    // On sufficiently small screens, we lock the workspaces panel to 80% of the screen
    // width. In this situation, we no longer need the panel to be resizable.
    return width < VIEWPORT_LG ? (
        <div className={styles.min80Width}>
            <Workspaces {...props} />
        </div>
    ) : (
        <Panel
            defaultSize={500}
            minSize={405}
            maxSize={0.45 * width}
            position="left"
            storageKey={WORKSPACES_LIST_SIZE}
            ariaLabel="execution workspaces"
        >
            <Workspaces {...props} />
        </Panel>
    )
}
