import React from 'react'

export interface BatchChangeChangesetsHeaderProps {
    enableSelect?: boolean
    allSelected?: boolean
    toggleSelectAll?: () => void
    disabled?: boolean
}

export const BatchChangeChangesetsHeader: React.FunctionComponent<BatchChangeChangesetsHeaderProps> = ({
    enableSelect,
    allSelected,
    toggleSelectAll,
    disabled,
}) => (
    <>
        <span className="d-none d-md-block" />
        {enableSelect && toggleSelectAll && (
            <input
                type="checkbox"
                className="btn ml-2"
                checked={allSelected}
                onChange={toggleSelectAll}
                disabled={!!disabled}
                data-tooltip="Click to select all changesets"
                aria-label="Click to select all changesets"
            />
        )}
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Status</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-nowrap">Changeset information</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Check state</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Review state</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Changes</h5>
    </>
)
