import React from 'react'

import {
    type HiddenExternalChangesetInfoCellProps,
    HiddenExternalChangesetInfoCell,
} from '../detail/changesets/HiddenExternalChangesetInfoCell'

import { ChangesetCloseActionKept } from './ChangesetCloseAction'

import styles from './HiddenExternalChangesetCloseNode.module.scss'

export interface HiddenExternalChangesetCloseNodeProps {
    node: HiddenExternalChangesetInfoCellProps['node']
}

export const HiddenExternalChangesetCloseNode: React.FunctionComponent<
    React.PropsWithChildren<HiddenExternalChangesetCloseNodeProps>
> = ({ node }) => (
    <>
        <span className="d-none d-sm-block" />
        {/* Hidden changesets are always untouched, so the action will always be "kept". */}
        <ChangesetCloseActionKept className={styles.hiddenExternalChangesetCloseNodeAction} />
        <HiddenExternalChangesetInfoCell node={node} className={styles.hiddenExternalChangesetCloseNodeInformation} />
        <span className="d-none d-sm-block" />
        <span className="d-none d-sm-block" />
        <span className="d-none d-sm-block" />
    </>
)
