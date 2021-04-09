import React from 'react'

export interface BatchChangeChangesetsHeaderProps {
    //  Nothing
}

export const BatchChangeChangesetsHeader: React.FunctionComponent<BatchChangeChangesetsHeaderProps> = () => (
    <>
        <span className="d-none d-md-block" />
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Status</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-nowrap">Changeset information</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Check state</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Review state</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Changes</h5>
    </>
)

export const BatchChangeChangesetsHeaderWithCheckboxes: React.FunctionComponent<BatchChangeChangesetsHeaderProps> = () => (
    <>
        <span className="d-none d-md-block" />
        <input
            id="select-all-changesets"
            type="checkbox"
            className="btn ml-2"
            // checked={selected}
            // onChange={toggleSelected}
            // disabled={!viewerCanAdminister}
            data-tooltip="Click to select all changesets"
        />
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Status</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-nowrap">Changeset information</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Check state</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Review state</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Changes</h5>
    </>
)
