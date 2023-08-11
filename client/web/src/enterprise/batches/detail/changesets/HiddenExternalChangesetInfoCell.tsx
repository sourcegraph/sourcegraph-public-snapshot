import React from 'react'

import classNames from 'classnames'

import { H3 } from '@sourcegraph/wildcard'

import type { HiddenExternalChangesetFields } from '../../../../graphql-operations'

import { ChangesetLastSynced } from './ChangesetLastSynced'

export interface HiddenExternalChangesetInfoCellProps {
    node: Pick<HiddenExternalChangesetFields, 'id' | 'nextSyncAt' | 'updatedAt' | '__typename'>
    className?: string
}

export const HiddenExternalChangesetInfoCell: React.FunctionComponent<
    React.PropsWithChildren<HiddenExternalChangesetInfoCellProps>
> = ({ node, className }) => (
    <div className={classNames('d-flex flex-column', className)}>
        <div className="m-0 mb-2">
            <H3 className="m-0 d-inline">
                <span className="text-muted">Changeset in a private repository</span>
            </H3>
        </div>
        <div>
            <ChangesetLastSynced changeset={node} viewerCanAdminister={false} />
        </div>
    </div>
)
