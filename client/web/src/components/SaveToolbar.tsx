import * as React from 'react'

import { mdiAlertCircle } from '@mdi/js'
import classNames from 'classnames'

import { Button, LoadingSpinner, Icon } from '@sourcegraph/wildcard'

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

export const SaveToolbar: React.FunctionComponent<
    React.PropsWithChildren<React.PropsWithChildren<SaveToolbarProps>>
> = ({ dirty, saving, error, onSave, onDiscard, children, willShowError, saveDiscardDisabled }) => {
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
                <div className={styles.error} role="alert">
                    <Icon className={styles.errorIcon} aria-hidden={true} svgPath={mdiAlertCircle} />
                    {error.message}
                </div>
            )}
            <div className={classNames('mt-2', styles.actions)}>
                <Button
                    disabled={disabled}
                    title={saveDiscardTitle || 'Save changes'}
                    className={classNames('test-save-toolbar-save mr-2', styles.item)}
                    onClick={onSave}
                    variant="primary"
                >
                    Save
                </Button>
                <Button
                    disabled={disabled}
                    title={saveDiscardTitle || 'Discard changes'}
                    className={classNames('test-save-toolbar-discard', styles.item)}
                    onClick={onDiscard}
                    variant="secondary"
                >
                    Discard changes
                </Button>
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
