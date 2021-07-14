import classNames from 'classnames'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import React, { useMemo, useCallback } from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { MultiSelectContext, MultiSelectContextState } from '../../MultiSelectContext'
import { BatchChangePreviewContext, BatchChangePreviewContextState } from '../BatchChangePreviewContext'

import styles from './PreviewListHeader.module.scss'

export interface PreviewListHeaderProps {
    selectionEnabled: boolean
}

export const PreviewListHeader: React.FunctionComponent<PreviewListHeaderProps> = ({ selectionEnabled }) => (
    <>
        {selectionEnabled && (
            <MultiSelectContext.Consumer>
                {selectProps => (
                    <BatchChangePreviewContext.Consumer>
                        {previewProps => <SelectAll {...selectProps} {...previewProps} />}
                    </BatchChangePreviewContext.Consumer>
                )}
            </MultiSelectContext.Consumer>
        )}
        <span className="p-2 d-none d-sm-block" />
        {selectionEnabled ? (
            <MultiSelectContext.Consumer>{props => <SelectVisibleCheckbox {...props} />}</MultiSelectContext.Consumer>
        ) : (
            <span className="d-none d-sm-block p-0 m-0" />
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
    Pick<MultiSelectContextState, 'selectVisible' | 'selectAll' | 'selected'> &
        Pick<BatchChangePreviewContextState, 'hasMorePages' | 'totalCount'>
> = ({ selectVisible, selectAll, selected, hasMorePages, totalCount }) => {
    const onClick = useCallback(() => {
        if (selected === 'all') {
            selectVisible()
        } else {
            selectAll()
        }
    }, [selected, selectVisible, selectAll])

    if (selected !== 'all' && selected.size === 0) {
        return null
    }

    return (
        <span className={classNames('m-0 p-0 d-none d-sm-block', styles.previewListHeaderSelectAll)}>
            <div className="ml-2 col d-flex align-items-center">
                <InfoCircleOutlineIcon className="icon-inline text-muted mr-2" />
                {selected === 'all'
                    ? `All ${totalCount} ${pluralize('changeset', totalCount)} selected`
                    : `${selected.size} ${pluralize('changeset', selected.size)} selected`}
                {hasMorePages && (
                    <button type="button" className="btn btn-link py-0 px-1" onClick={onClick}>
                        ({selected === 'all' ? 'Deselect' : 'Select'} all {totalCount})
                    </button>
                )}
            </div>
        </span>
    )
}

const SelectVisibleCheckbox: React.FunctionComponent<
    Pick<MultiSelectContextState, 'deselectVisible' | 'selectVisible' | 'selected' | 'visible'>
> = ({ deselectVisible: deselectVisible, selectVisible: selectVisible, selected, visible }) => {
    const checked = useMemo(() => selected === 'all' || selected.size === visible.size, [selected, visible])
    const disabled = useMemo(() => selected === 'all', [selected])
    const onChange = useCallback(() => {
        if (checked) {
            deselectVisible()
        } else {
            selectVisible()
        }
    }, [checked, deselectVisible, selectVisible])

    return (
        <span className="p-2 pl-3 d-none d-sm-block">
            <input
                type="checkbox"
                checked={checked}
                disabled={disabled}
                onChange={onChange}
                data-tooltip={`Click to ${checked ? 'deselect' : 'select'} all visible changesets`}
            />
        </span>
    )
}
