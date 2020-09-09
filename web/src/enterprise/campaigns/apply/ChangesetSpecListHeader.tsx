import React from 'react'

export interface ChangesetSpecListHeaderProps {
    // Nothing for now.
}

export const ChangesetSpecListHeader: React.FunctionComponent<ChangesetSpecListHeaderProps> = () => (
    <>
        <span />
        <h5 className="text-uppercase text-center text-nowrap text-muted">Action</h5>
        <h5 className="text-uppercase text-nowrap text-muted">Changeset information</h5>
        <h5 className="text-uppercase text-right text-nowrap text-muted">Changes</h5>
    </>
)
