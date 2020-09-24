import React from 'react'
import { HiddenExternalChangesetFields } from '../../../../graphql-operations'
import { ChangesetLastSynced } from './ChangesetLastSynced'

export interface HiddenExternalChangesetInfoCellProps {
    node: Pick<HiddenExternalChangesetFields, 'id' | 'nextSyncAt' | 'updatedAt'>
}

export const HiddenExternalChangesetInfoCell: React.FunctionComponent<HiddenExternalChangesetInfoCellProps> = ({
    node,
}) => (
    <div className="d-flex flex-column">
        <div className="m-0 mb-2">
            <h3 className="m-0 d-inline">
                <span className="text-muted">Changeset in a private repository</span>
            </h3>
        </div>
        <div>
            <ChangesetLastSynced changeset={node} viewerCanAdminister={false} />
        </div>
    </div>
)
