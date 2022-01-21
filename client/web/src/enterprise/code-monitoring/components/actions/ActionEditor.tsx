import classNames from 'classnames'
import React, { useCallback, useState } from 'react'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { Button } from '@sourcegraph/wildcard'

import styles from '../CodeMonitorForm.module.scss'

interface Props {
    title: React.ReactNode
    subtitle: string
    disabled: boolean
    completed: boolean
    completedSubtitle: string

    actionEnabled: boolean
    toggleActionEnabled: (enabled: boolean) => void

    onSubmit: React.FormEventHandler
    onCancel?: React.FormEventHandler

    canDelete: boolean
    onDelete: React.FormEventHandler
}

export const ActionEditor: React.FunctionComponent<Props> = ({
    title,
    subtitle,
    disabled,
    completed,
    completedSubtitle,
    actionEnabled,
    toggleActionEnabled,
    onSubmit,
    onCancel,
    canDelete,
    onDelete,
    children,
}) => {
    const [expanded, setExpanded] = useState(false)
    const toggleExpanded = useCallback(() => {
        if (!disabled) {
            setExpanded(expanded => !expanded)
        }
    }, [disabled])

    const submitHandler: React.FormEventHandler = useCallback(
        event => {
            setExpanded(false)
            onSubmit(event)
        },
        [onSubmit]
    )

    const cancelHandler: React.FormEventHandler = useCallback(
        event => {
            setExpanded(false)
            if (onCancel) {
                onCancel(event)
            }
        },
        [onCancel]
    )

    const deleteHandler: React.FormEventHandler = useCallback(
        event => {
            setExpanded(false)
            onDelete(event)
        },
        [onDelete]
    )

    // When the action is completed, the wrapper cannot be a button because we show nested buttons inside it.
    // Use a div instead. The edit button will still allow keyboard users to activate the form.
    const CollapsedWrapperElement = completed ? 'div' : Button

    return (
        <>
            {expanded && (
                <div className={classNames(styles.card, 'card p-3')}>
                    <div className="font-weight-bold">{title}</div>
                    <span className="text-muted">{subtitle}</span>

                    {children}

                    <div className="d-flex align-items-center my-4">
                        <div>
                            <Toggle
                                title="Enabled"
                                value={actionEnabled}
                                onToggle={toggleActionEnabled}
                                className="mr-2"
                                aria-labelledby="code-monitoring-form-actions-enable-toggle"
                            />
                        </div>
                        <span id="code-monitoring-form-actions-enable-toggle">
                            {actionEnabled ? 'Enabled' : 'Disabled'}
                        </span>
                    </div>
                    <div className="d-flex justify-content-between">
                        <div>
                            <Button
                                type="submit"
                                data-testid="submit-action"
                                className="mr-1 test-submit-action"
                                onClick={submitHandler}
                                onSubmit={submitHandler}
                                variant="secondary"
                            >
                                Continue
                            </Button>
                            <Button onClick={cancelHandler} outline={true} variant="secondary">
                                Cancel
                            </Button>
                        </div>
                        {canDelete && (
                            <Button onClick={deleteHandler} outline={true} variant="danger">
                                Delete
                            </Button>
                        )}
                    </div>
                </div>
            )}
            {!expanded && (
                <CollapsedWrapperElement
                    data-testid="form-action-toggle-email-notification"
                    className={classNames('card test-action-button', styles.cardButton, disabled && 'disabled')}
                    disabled={disabled}
                    aria-label="Edit action: Send email notifications"
                    onClick={toggleExpanded}
                >
                    <div className="d-flex justify-content-between align-items-center w-100">
                        <div>
                            <div
                                className={classNames(
                                    'font-weight-bold',
                                    !completed && classNames(styles.cardLink, 'btn-link')
                                )}
                            >
                                {title}
                            </div>
                            {completed ? (
                                <span className="text-muted font-weight-normal" data-testid="existing-action-email">
                                    {completedSubtitle}
                                </span>
                            ) : (
                                <span className="text-muted font-weight-normal">{subtitle}</span>
                            )}
                        </div>
                        {completed && (
                            <div className="d-flex align-items-center">
                                <div>
                                    <Toggle
                                        title="Enabled"
                                        value={actionEnabled}
                                        onToggle={toggleActionEnabled}
                                        className="mr-3"
                                    />
                                </div>
                                <Button variant="link" className="p-0">
                                    Edit
                                </Button>
                            </div>
                        )}
                    </div>
                </CollapsedWrapperElement>
            )}
        </>
    )
}
