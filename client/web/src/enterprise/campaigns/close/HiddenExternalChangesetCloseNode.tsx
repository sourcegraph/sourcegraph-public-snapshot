import React from 'react'
import { ChangesetCloseActionKept } from './ChangesetCloseAction'
import {
    HiddenExternalChangesetInfoCellProps,
    HiddenExternalChangesetInfoCell,
} from '../detail/changesets/HiddenExternalChangesetInfoCell'

export interface HiddenExternalChangesetCloseNodeProps {
    node: HiddenExternalChangesetInfoCellProps['node']
}

export const HiddenExternalChangesetCloseNode: React.FunctionComponent<HiddenExternalChangesetCloseNodeProps> = ({
    node,
}) => (
    <>
        <span />
        {/* Hidden changesets are always untouched, so the action will always be "kept". */}
        <ChangesetCloseActionKept />
        <HiddenExternalChangesetInfoCell node={node} />
        <span />
        <span />
        <span />
    </>
)
