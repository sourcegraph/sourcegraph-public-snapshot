import React, { useState, useCallback } from 'react'
import { Toggle } from '../../../../../branded/src/components/Toggle'
import { AuthenticatedUser } from '../../../auth'
import { CodeMonitorFields } from '../../../graphql-operations'

interface ActionAreaProps {
    actions: CodeMonitorFields['actions']
    actionsCompleted: boolean
    setActionsCompleted: (completed: boolean) => void
    disabled: boolean
    authenticatedUser: AuthenticatedUser
    onActionsChange: (action: CodeMonitorFields['actions']) => void
}

export const FormActionArea: React.FunctionComponent<ActionAreaProps> = ({
    actions,
    actionsCompleted,
    setActionsCompleted,
    disabled,
    authenticatedUser,
    onActionsChange,
}) => {
    const [showEmailNotificationForm, setShowEmailNotificationForm] = useState(false)
    const toggleEmailNotificationForm: React.FormEventHandler = useCallback(event => {
        event.preventDefault()
        setShowEmailNotificationForm(show => !show)
    }, [])

    const [emailNotificationEnabled, setEmailNotificationEnabled] = useState(true)
    const toggleEmailNotificationEnabled: (enabled: boolean) => void = useCallback(
        enabled => {
            setEmailNotificationEnabled(enabled)
            onActionsChange({ nodes: [{ id: '', recipients: { nodes: [{ id: authenticatedUser.email }] }, enabled }] })
        },
        [authenticatedUser, onActionsChange]
    )

    const completeForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setShowEmailNotificationForm(false)
            setActionsCompleted(true)
            if (actions.nodes.length === 0) {
                // ID can be empty here, since we'll generate a new ID when we create the monitor.
                onActionsChange({
                    nodes: [{ id: '', enabled: true, recipients: { nodes: [{ id: authenticatedUser.id }] } }],
                })
            }
        },
        [setActionsCompleted, setShowEmailNotificationForm, actions.nodes.length, authenticatedUser.id, onActionsChange]
    )
    const editForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setShowEmailNotificationForm(true)
        },
        [setShowEmailNotificationForm]
    )
    const cancelForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setShowEmailNotificationForm(false)
        },
        [setShowEmailNotificationForm]
    )

    return (
        <>
            <h3 className="mb-1">Actions</h3>
            <span className="text-muted">Run any number of actions in response to an event</span>
            {/* This should be its own component when you can add multiple email actions */}
            {!showEmailNotificationForm && !actionsCompleted && (
                <button
                    type="button"
                    className="code-monitor-form__card card p-3 my-3 w-100 test-action-button"
                    onClick={toggleEmailNotificationForm}
                    disabled={disabled}
                >
                    <div className="code-monitor-form__card-link btn btn-link font-weight-bold p-0 text-left">
                        Send email notifications
                    </div>
                    <span className="text-muted">Deliver email notifications to specified recipients.</span>
                </button>
            )}
            {showEmailNotificationForm && (
                <div className="code-monitor-form__card card p-3 my-3">
                    <div className="font-weight-bold">Send email notifications</div>
                    <span className="text-muted">Deliver email notifications to specified recipients.</span>
                    <div className="mt-4 test-action-form">
                        Recipients
                        <input
                            type="text"
                            className="form-control my-2"
                            value={`${authenticatedUser.email || ''} (you)`}
                            disabled={true}
                            autoFocus={true}
                        />
                        <small className="text-muted">
                            Code monitors are currently limited to sending emails to your primary email address.
                        </small>
                    </div>
                    <div className="flex">
                        <Toggle
                            title="Enabled"
                            value={emailNotificationEnabled}
                            onToggle={toggleEmailNotificationEnabled}
                            className="mr-2"
                        />
                        Enabled
                    </div>
                    <div>
                        <button
                            type="submit"
                            className="btn btn-secondary mr-1 test-submit-action"
                            onClick={completeForm}
                            onSubmit={completeForm}
                        >
                            Done
                        </button>
                        <button type="button" className="btn btn-outline-secondary" onClick={cancelForm}>
                            Cancel
                        </button>
                    </div>
                </div>
            )}
            {actionsCompleted && !showEmailNotificationForm && (
                <div className="code-monitor-form__card card p-3 my-3">
                    <div className="d-flex justify-content-between align-items-center">
                        <div>
                            <div className="font-weight-bold">Send email notifications</div>
                            <span className="text-muted test-existing-action-email">{authenticatedUser.email}</span>
                        </div>
                        <div className="d-flex">
                            <div className="flex my-4">
                                <Toggle
                                    title="Enabled"
                                    value={emailNotificationEnabled}
                                    onToggle={toggleEmailNotificationEnabled}
                                    className="mr-2"
                                />
                            </div>
                            <button
                                type="button"
                                onClick={editForm}
                                className="btn btn-link p-0 text-left test-edit-action"
                            >
                                Edit
                            </button>
                        </div>
                    </div>
                </div>
            )}
            <small className="text-muted">
                What other actions would you like to do?{' '}
                <a href="" target="_blank" rel="noopener">
                    {/* TODO: populate link */}
                    Share feedback.
                </a>
            </small>
        </>
    )
}
