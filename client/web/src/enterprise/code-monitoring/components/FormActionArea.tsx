import React, { useState, useCallback } from 'react'
import { Toggle } from '../../../../../branded/src/components/Toggle'
import { AuthenticatedUser } from '../../../auth'
import { Action } from './CodeMonitorForm'

interface ActionAreaProps {
    actionCompleted: boolean
    setActionCompleted: () => void
    disabled: boolean
    authenticatedUser: AuthenticatedUser
    onActionChange: (action: Action) => void
}

export const FormActionArea: React.FunctionComponent<ActionAreaProps> = ({
    actionCompleted,
    setActionCompleted,
    disabled,
    authenticatedUser,
    onActionChange,
}) => {
    const [showEmailNotificationForm, setShowEmailNotificationForm] = useState(false)
    const toggleEmailNotificationForm: React.FormEventHandler = useCallback(event => {
        event.preventDefault()
        setShowEmailNotificationForm(show => !show)
    }, [])

    const editOrCompleteForm: React.FormEventHandler = useCallback(
        event => {
            event?.preventDefault()
            toggleEmailNotificationForm(event)
            setActionCompleted()
        },
        [toggleEmailNotificationForm, setActionCompleted]
    )

    const [emailNotificationEnabled, setEmailNotificationEnabled] = useState(true)
    const toggleEmailNotificationEnabled: (value: boolean) => void = useCallback(
        enabled => {
            setEmailNotificationEnabled(enabled)
            onActionChange({ recipient: authenticatedUser.email, enabled })
        },
        [authenticatedUser, onActionChange]
    )

    return (
        <>
            <h3 className="mb-1">Actions</h3>
            <span className="text-muted">Run any number of actions in response to an event</span>
            <div className="card p-3 my-3">
                {/* This should be its own component when you can add multiple email actions */}
                {!showEmailNotificationForm && !actionCompleted && (
                    <>
                        <button
                            type="button"
                            onClick={toggleEmailNotificationForm}
                            className="btn btn-link font-weight-bold p-0 text-left test-action-button"
                            disabled={disabled}
                        >
                            Send email notifications
                        </button>
                        <span className="text-muted">Deliver email notifications to specified recipients.</span>
                    </>
                )}
                {showEmailNotificationForm && !actionCompleted && (
                    <>
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
                        <div className="flex my-4">
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
                                className="btn btn-outline-secondary mr-1 test-submit-action"
                                onClick={editOrCompleteForm}
                                onSubmit={editOrCompleteForm}
                            >
                                Done
                            </button>
                            <button type="button" className="btn btn-outline-secondary" onClick={editOrCompleteForm}>
                                Cancel
                            </button>
                        </div>
                    </>
                )}
                {actionCompleted && (
                    <div className="d-flex justify-content-between align-items-center">
                        <div>
                            <div className="font-weight-bold">Send email notifications</div>
                            <span className="text-muted">{authenticatedUser.email}</span>
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
                            <button type="button" onClick={editOrCompleteForm} className="btn btn-link p-0 text-left">
                                Edit
                            </button>
                        </div>
                    </div>
                )}
            </div>
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
