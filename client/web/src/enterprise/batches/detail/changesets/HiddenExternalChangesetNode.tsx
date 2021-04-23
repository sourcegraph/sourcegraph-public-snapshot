import classNames from 'classnames'
import React from 'react'

import { HiddenExternalChangesetFields } from '../../../../graphql-operations'

import { ChangesetStatusCell } from './ChangesetStatusCell'
import { HiddenExternalChangesetInfoCell } from './HiddenExternalChangesetInfoCell'
import styles from './HiddenExternalChangesetNode.module.scss'

export interface HiddenExternalChangesetNodeProps {
    node: Pick<HiddenExternalChangesetFields, 'id' | 'nextSyncAt' | 'updatedAt' | 'state' | '__typename'>
}

export const HiddenExternalChangesetNode: React.FunctionComponent<HiddenExternalChangesetNodeProps> = ({ node }) => (
    <>
        <span className="d-none d-sm-block" />
        <div className="p-2">
            <input
                id={`select-changeset-${node.id}`}
                type="checkbox"
                className="btn"
                checked={false}
                disabled={true}
                data-tooltip="You do not have permission to detach this changeset"
            />
        </div>
        <ChangesetStatusCell
            id={node.id}
            state={node.state}
            className={classNames(styles.hiddenExternalChangesetNodeStatus, 'p-2 text-muted d-block d-sm-flex')}
        />
        <HiddenExternalChangesetInfoCell
            node={node}
            className={classNames(styles.hiddenExternalChangesetNodeInformation, 'p-2')}
        />
        <span className="d-none d-sm-block" />
        <span className="d-none d-sm-block" />
        <span className="d-none d-sm-block" />
    </>
)
