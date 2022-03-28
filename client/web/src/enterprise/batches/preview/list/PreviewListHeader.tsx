import React from 'react'

import { InputTooltip } from '../../../../components/InputTooltip'

export interface PreviewListHeaderProps {
    allSelected?: boolean
    toggleSelectAll?: () => void
}

export const PreviewListHeader: React.FunctionComponent<PreviewListHeaderProps> = ({
    allSelected,
    toggleSelectAll,
}) => (
    <>
        <span className="p-2 d-none d-sm-block" />
        {toggleSelectAll && (
            <div className="d-flex p-2 align-items-center">
                <InputTooltip
                    type="checkbox"
                    checked={allSelected}
                    onChange={toggleSelectAll}
                    tooltip="Click to select all changesets"
                    aria-label="Click to select all changesets"
                />
                <span className="pl-2 d-block d-sm-none">Select all</span>
            </div>
        )}
        <h5 className="p-2 d-none d-sm-block text-uppercase text-center">Current state</h5>
        <h5 className="d-none d-sm-block text-uppercase text-center">
            +<br />-
        </h5>
        <h5 className="p-2 d-none d-sm-block text-uppercase text-nowrap">Actions</h5>
        <h5 className="p-2 d-none d-sm-block text-uppercase text-nowrap">Changeset information</h5>
        <h5 className="p-2 d-none d-sm-block text-uppercase text-center text-nowrap">Commit changes</h5>
        <h5 className="p-2 d-none d-sm-block text-uppercase text-center text-nowrap">Change state</h5>
    </>
)
