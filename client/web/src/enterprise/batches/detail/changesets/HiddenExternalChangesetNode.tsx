import React from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'

import { InputTooltip } from '../../../../components/InputTooltip'
import type { HiddenExternalChangesetFields } from '../../../../graphql-operations'

import { ChangesetStatusCell } from './ChangesetStatusCell'
import { HiddenExternalChangesetInfoCell } from './HiddenExternalChangesetInfoCell'

import styles from './HiddenExternalChangesetNode.module.scss'

export interface HiddenExternalChangesetNodeProps {
    node: Pick<HiddenExternalChangesetFields, 'id' | 'nextSyncAt' | 'updatedAt' | 'state' | '__typename'>
}

export const HiddenExternalChangesetNode: React.FunctionComponent<
    React.PropsWithChildren<HiddenExternalChangesetNodeProps>
> = ({ node }) => (
    <>
        <span className="d-none d-sm-block" />
        <div className="p-2">
            {/* eslint-disable-next-line no-restricted-syntax*/}
            <InputTooltip
                id={`select-changeset-${node.id}`}
                type="checkbox"
                checked={false}
                disabled={true}
                tooltip="You do not have permission to perform a bulk operation on this changeset"
                aria-label="You do not have permission to perform a bulk operation on this changeset"
                placement="right"
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
        <span className="d-none d-sm-block">
            <VisuallyHidden>Check state, review state, and diff unavailable</VisuallyHidden>
        </span>
        <span className="d-none d-sm-block" />
        <span className="d-none d-sm-block" />
    </>
)
