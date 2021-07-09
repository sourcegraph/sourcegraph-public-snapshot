import React, { useMemo, useCallback } from 'react'

import { MultiSelectContext, MultiSelectContextState } from '../../MultiSelectContext'

export interface PreviewListHeaderProps {
    selectionEnabled: boolean
}

export const PreviewListHeader: React.FunctionComponent<PreviewListHeaderProps> = ({ selectionEnabled }) => (
    <>
        <span className="p-2 d-none d-sm-block" />
        {selectionEnabled && (
            <MultiSelectContext.Consumer>{props => <SelectAllCheckbox {...props} />}</MultiSelectContext.Consumer>
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

const SelectAllCheckbox: React.FunctionComponent<
    Pick<MultiSelectContextState, 'onDeselectAll' | 'onSelectAll' | 'selected'>
> = ({ onDeselectAll, onSelectAll, selected }) => {
    const checked = useMemo(() => selected === 'all', [selected])
    const onClick = useCallback(() => {
        if (checked) {
            onDeselectAll()
        } else {
            onSelectAll()
        }
    }, [checked, onDeselectAll, onSelectAll])

    return (
        <span className="p-2 d-none d-sm-block">
            <input type="checkbox" checked={checked} onClick={onClick} />
        </span>
    )
}
