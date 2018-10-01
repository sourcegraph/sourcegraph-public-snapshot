import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
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
        <>
            {error &&
                !saving && (
                    <div className="save-toolbar__error">
                        <AlertCircleIcon className="icon-inline save-toolbar__error-icon" />
                        {error.message}
                    </div>
                )}
            <div className="save-toolbar__actions">
                <button
                    disabled={saveDiscardDisabled}
                    title={saveDiscardTitle || 'Save changes'}
                    className="btn btn-sm btn-success save-toolbar__item save-toolbar__btn save-toolbar__btn-first"
                    onClick={onSave}
                >
                    <CheckIcon className="icon-inline" /> Save changes
                </button>
                <button
                    disabled={saveDiscardDisabled}
                    title={saveDiscardTitle || 'Discard changes'}
                    className="btn btn-sm btn-secondary save-toolbar__item save-toolbar__btn save-toolbar__btn-last"
                    onClick={onDiscard}
                >
                    <CloseIcon className="icon-inline" /> Discard
                </button>
                {saving && (
                    <span className="save-toolbar__item save-toolbar__message">
                        <LoadingSpinner className="icon-inline" /> Saving...
                    </span>
                )}
            </div>
        </>
    )
}
