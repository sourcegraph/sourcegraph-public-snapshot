import React from 'react'

export interface ChangesetSpecListHeaderProps {
    // Nothing for now.
}

export const ChangesetSpecListHeader: React.FunctionComponent<ChangesetSpecListHeaderProps> = () => (
    <>
        <span />
        <h5 className="text-uppercase text-center text-nowrap">Action</h5>
        <h5 className="text-uppercase text-nowrap">Changeset information</h5>
        <h5 className="text-uppercase text-center text-nowrap">Changes</h5>
    </>
)
