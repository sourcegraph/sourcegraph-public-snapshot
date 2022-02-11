import classNames from 'classnames'
import React, { useCallback, useState } from 'react'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { Button, Card } from '@sourcegraph/wildcard'

import styles from '../CodeMonitorForm.module.scss'

export interface ActionEditorProps {
    title: React.ReactNode
    label: string // Similar to title, but for string-only labels
    subtitle: string
    disabled: boolean
    completed: boolean
    completedSubtitle: string
    idName: string // Name used for generating IDs, including form control IDs and test IDs

    actionEnabled: boolean
    toggleActionEnabled: (enabled: boolean) => void

    canSubmit?: boolean
    onSubmit: React.FormEventHandler
    onCancel?: React.FormEventHandler

    canDelete: boolean
    onDelete: React.FormEventHandler

    // For testing purposes only
    _testStartOpen?: boolean
}

export const ActionEditor: React.FunctionComponent<ActionEditorProps> = ({
    title,
    label,
    subtitle,
    disabled,
    completed,
    completedSubtitle,
    idName,
    actionEnabled,
    toggleActionEnabled,
    canSubmit = true,
    onSubmit,
    onCancel,
    canDelete,
    onDelete,
    children,
    _testStartOpen = false,
}) => {
    const [expanded, setExpanded] = useState(_testStartOpen)
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

    return (
        <>
            {expanded && (
                <Card className={classNames(styles.card, 'p-3')}>
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
                                aria-labelledby={`code-monitoring-${idName}-form-actions-enable-toggle`}
                                data-testid={`enable-action-toggle-expanded-${idName}`}
                            />
                        </div>
                        <span id={`code-monitoring-${idName}-form-actions-enable-toggle`}>
                            {actionEnabled ? 'Enabled' : 'Disabled'}
                        </span>
                    </div>
                    <div className="d-flex justify-content-between">
                        <div>
                            <Button
                                data-testid={`submit-action-${idName}`}
                                className={`mr-1 test-submit-action-${idName}`}
                                onClick={submitHandler}
                                disabled={!canSubmit}
                                variant="secondary"
                            >
                                Continue
                            </Button>
                            <Button
                                onClick={cancelHandler}
                                outline={true}
                                variant="secondary"
                                data-testid={`cancel-action-${idName}`}
                            >
                                Cancel
                            </Button>
                        </div>
                        {canDelete && (
                            <Button
                                onClick={deleteHandler}
                                outline={true}
                                variant="danger"
                                data-testid={`delete-action-${idName}`}
                            >
                                Delete
                            </Button>
                        )}
                    </div>
                </Card>
            )}
            {!expanded && (
                <Card
                    // When the action is completed, the wrapper cannot be a button because we show nested buttons inside it.
                    // Use a div instead. The edit button will still allow keyboard users to activate the form.
                    as={completed ? 'div' : Button}
                    data-testid={`form-action-toggle-${idName}`}
                    className={classNames(
                        `test-action-button-${idName}`,
                        styles.cardButton,
                        disabled && styles.btnDisabled
                    )}
                    disabled={disabled}
                    aria-label={`Edit action: ${label}`}
                    onClick={toggleExpanded}
                >
                    <div className="d-flex justify-content-between align-items-center w-100">
                        <div>
                            <div className={classNames('font-weight-bold', !completed && styles.cardLink)}>{title}</div>
                            {completed ? (
                                <span
                                    className="text-muted font-weight-normal"
                                    data-testid={`existing-action-${idName}`}
                                >
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
                                        data-testid={`enable-action-toggle-collapsed-${idName}`}
                                    />
                                </div>
                                <Button variant="link" className="p-0">
                                    Edit
                                </Button>
                            </div>
                        )}
                    </div>
                </Card>
            )}
        </>
    )
}
