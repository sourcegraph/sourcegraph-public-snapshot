import React from 'react'
import { ChangesetSpecFields } from '../../../graphql-operations'

export interface ChangesetSpecListHeaderProps {
    nodes: ChangesetSpecFields[]
    totalCount?: number | null
}

export const ChangesetSpecListHeader: React.FunctionComponent<ChangesetSpecListHeaderProps> = ({
    nodes,
    totalCount,
}) => (
    <>
        <div className="changeset-spec-list-header__title mb-2">
            <strong>
                Displaying {nodes.length}
                {totalCount && <> of {totalCount}</>} changesets
            </strong>
        </div>
        <span />
        <h5 className="text-uppercase text-center text-nowrap text-muted">Action</h5>
        <h5 className="text-uppercase text-nowrap text-muted">Changeset information</h5>
        <h5 className="text-uppercase text-right text-nowrap text-muted">Changes</h5>
    </>
)
