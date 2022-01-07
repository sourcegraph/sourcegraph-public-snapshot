import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'

import { LoadingSpinner } from '@sourcegraph/wildcard'

import styles from './SaveToolbar.module.scss'

export interface SaveToolbarProps {
    dirty?: boolean
    saving?: boolean
    error?: Error

    onSave: () => void
    onDiscard: () => void
    /**
     * Override under what circumstances the error can be shown.
     * Setting this does not include the default behavior of `return !saving`
     */
    willShowError?: () => boolean
    /**
     * Override under what circumstances the save and discard buttons will be disabled.
     * Setting this does not include the default behavior of `return !saving`
     */
    saveDiscardDisabled?: () => boolean
}

export type SaveToolbarPropsGenerator<T extends object> = (
    props: Readonly<React.PropsWithChildren<SaveToolbarProps>>
) => React.PropsWithChildren<SaveToolbarProps> & T

export const SaveToolbar: React.FunctionComponent<React.PropsWithChildren<SaveToolbarProps>> = ({
    dirty,
    saving,
    error,
    onSave,
    onDiscard,
    children,
    willShowError,
    saveDiscardDisabled,
}) => {
    const disabled = saveDiscardDisabled ? saveDiscardDisabled() : saving || !dirty
    let saveDiscardTitle: string | undefined
    if (saving) {
        saveDiscardTitle = 'Saving...'
    } else if (!dirty) {
        saveDiscardTitle = 'No changes to save or discard'
    }

    if (!willShowError) {
        willShowError = (): boolean => !saving
    }

    return (
        <>
            {error && willShowError() && (
                <div className={styles.error}>
                    <AlertCircleIcon className={classNames('icon-inline', styles.errorIcon)} />
                    {error.message}
                </div>
            )}
            <div className={styles.actions}>
                <button
                    type="button"
                    disabled={disabled}
                    title={saveDiscardTitle || 'Save changes'}
                    className={classNames(
                        'btn btn-sm btn-success test-save-toolbar-save',
                        styles.item,
                        styles.btn,
                        styles.btnFirst
                    )}
                    onClick={onSave}
                >
                    <CheckIcon className="icon-inline" style={{ marginRight: '0.1em' }} /> Save changes
                </button>
                <button
                    type="button"
                    disabled={disabled}
                    title={saveDiscardTitle || 'Discard changes'}
                    className={classNames(
                        'btn btn-sm btn-secondary test-save-toolbar-discard',
                        styles.item,
                        styles.btn,
                        styles.btnLast
                    )}
                    onClick={onDiscard}
                >
                    <CloseIcon className="icon-inline" /> Discard
                </button>
                {children}
                {saving && (
                    <span className={classNames(styles.item, styles.message)}>
                        <LoadingSpinner /> Saving...
                    </span>
                )}
            </div>
        </>
    )
}
