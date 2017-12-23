import CheckmarkIcon from '@sourcegraph/icons/lib/Checkmark'
import CloseIcon from '@sourcegraph/icons/lib/Close'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import Loader from '@sourcegraph/icons/lib/Loader'
import * as React from 'react'

interface Props {
    dirty?: boolean
    disabled?: boolean
    saving?: boolean
    error?: Error

    onSave: () => void
    onDiscard: () => void
}

export const SaveToolbar: React.SFC<Props> = ({ dirty, disabled, saving, error, onSave, onDiscard }) => {
    const saveDiscardDisabled = saving || !dirty
    let saveDiscardTitle: string | undefined
    if (saving) {
        saveDiscardTitle = 'Saving...'
    } else if (!dirty) {
        saveDiscardTitle = 'No changes to save or discard'
    }

    return (
        <div>
            <div className="save-toolbar__actions">
                <button
                    disabled={saveDiscardDisabled}
                    title={saveDiscardTitle || 'Save changes'}
                    className="btn btn-sm btn-link save-toolbar__action"
                    onClick={onSave}
                >
                    <CheckmarkIcon className="icon-inline" /> Save
                </button>
                <button
                    disabled={saveDiscardDisabled}
                    title={saveDiscardTitle || 'Discard changes'}
                    className="btn btn-sm btn-link save-toolbar__action"
                    onClick={onDiscard}
                >
                    <CloseIcon className="icon-inline" /> Discard
                </button>
                {saving && (
                    <span className="save-toolbar__action">
                        <Loader className="icon-inline" /> Saving...
                    </span>
                )}
            </div>
            {error && (
                <div className="save-toolbar__error">
                    <ErrorIcon className="icon-inline save-toolbar__error-icon" />
                    {error.message}
                </div>
            )}
        </div>
    )
}
