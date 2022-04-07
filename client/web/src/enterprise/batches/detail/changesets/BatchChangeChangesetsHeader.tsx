import React from 'react'

import { InputTooltip } from '../../../../components/InputTooltip'

export interface BatchChangeChangesetsHeaderProps {
    allSelected?: boolean
    toggleSelectAll?: () => void
    disabled?: boolean
}

export const BatchChangeChangesetsHeader: React.FunctionComponent<BatchChangeChangesetsHeaderProps> = ({
    allSelected,
    toggleSelectAll,
    disabled,
}) => (
    <>
        <span className="d-none d-md-block" />
        {toggleSelectAll && (
            <InputTooltip
                type="checkbox"
                className="ml-2"
                checked={allSelected}
                onChange={toggleSelectAll}
                disabled={!!disabled}
                tooltip={
                    disabled ? 'You do not have permission to perform this operation' : 'Click to select all changesets'
                }
                aria-label={
                    disabled ? 'You do not have permission to perform this operation' : 'Click to select all changesets'
                }
            />
        )}
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Status</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-nowrap">Changeset information</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Check state</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Review state</h5>
        <h5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Changes</h5>
    </>
)
