import classNames from 'classnames'
import React, { useMemo, useCallback } from 'react'

import { MultiSelectContext, MultiSelectContextState } from '../../MultiSelectContext'

import styles from './PreviewListHeader.module.scss'

export interface PreviewListHeaderProps {
    selectionEnabled: boolean
}

export const PreviewListHeader: React.FunctionComponent<PreviewListHeaderProps> = ({ selectionEnabled }) => (
    <>
        {selectionEnabled && (
            <MultiSelectContext.Consumer>{props => <SelectAll {...props} />}</MultiSelectContext.Consumer>
        )}
        <span className="p-2 d-none d-sm-block" />
        {selectionEnabled && (
            <MultiSelectContext.Consumer>{props => <SelectVisibleCheckbox {...props} />}</MultiSelectContext.Consumer>
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

const SelectAll: React.FunctionComponent<
    Pick<MultiSelectContextState, 'onSelectVisible' | 'onSelectAll' | 'selected' | 'visible'>
> = ({ onSelectVisible, onSelectAll, selected, visible }) => {
    const isVisible = useMemo(() => selected === 'all' || selected.size === visible.size, [selected, visible])
    const onClick = useCallback(() => {
        if (selected === 'all') {
            onSelectVisible()
        } else {
            onSelectAll()
        }
    }, [selected, onSelectVisible, onSelectAll])

    if (!isVisible) {
        return null
    }

    return (
        <span className={classNames('p-2 d-none d-sm-block', styles.previewListHeaderSelectAll)}>
            <button type="button" className="btn btn-link" onClick={onClick}>
                {selected === 'all' ? 'Deselect' : 'Select'} all changesets
            </button>
        </span>
    )
}

const SelectVisibleCheckbox: React.FunctionComponent<
    Pick<MultiSelectContextState, 'onDeselectVisible' | 'onSelectVisible' | 'selected' | 'visible'>
> = ({ onDeselectVisible, onSelectVisible, selected, visible }) => {
    const checked = useMemo(() => selected === 'all' || selected.size === visible.size, [selected, visible])
    const disabled = useMemo(() => selected === 'all', [selected])
    const onClick = useCallback(() => {
        if (checked) {
            onDeselectVisible()
        } else {
            onSelectVisible()
        }
    }, [checked, onDeselectVisible, onSelectVisible])

    return (
        <span className="p-2 pl-3 d-none d-sm-block">
            <input
                type="checkbox"
                checked={checked}
                disabled={disabled}
                onClick={onClick}
                data-tooltip={`Click to ${checked ? 'deselect' : 'select'} all visible changesets`}
            />
        </span>
    )
}
