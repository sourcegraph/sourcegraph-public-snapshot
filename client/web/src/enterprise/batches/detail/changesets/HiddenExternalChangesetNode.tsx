import React from 'react'

import classNames from 'classnames'

import { Button, ButtonProps, Tooltip } from '@sourcegraph/wildcard'

import { HiddenExternalChangesetFields } from '../../../../graphql-operations'

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
            <Tooltip
                content="You do not have permission to perform a bulk operation on this changeset"
                placement="right"
            >
                <Button
                    as="input"
                    type={'checkbox' as ButtonProps['type']}
                    aria-label="You do not have permission to perform a bulk operation on this changeset"
                    id={`select-changeset-${node.id}`}
                    checked={false}
                    disabled={true}
                />
            </Tooltip>
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
