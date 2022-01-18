import classNames from 'classnames'
import React, { useState, useCallback, useEffect } from 'react'
import { Observable } from 'rxjs'
import { delay, startWith, tap, mergeMap, catchError } from 'rxjs/operators'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { asError, isErrorLike } from '@sourcegraph/common'
import { useEventObservable, Button } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { MonitorEmailPriority } from '../../../../graphql-operations'
import { triggerTestEmailAction } from '../../backend'
import { MonitorAction } from '../FormActionArea'
import styles from '../FormActionArea.module.scss'

const LOADING = 'LOADING' as const

interface EmailActionProps {
    action?: MonitorAction
    setAction: (action?: MonitorAction) => void
    actionCompleted: boolean
    setActionCompleted: (actionCompleted: boolean) => void
    disabled: boolean
    authenticatedUser: AuthenticatedUser
    description: string
    cardClassName?: string
    cardBtnClassName?: string
    cardLinkClassName?: string
}

export const EmailAction: React.FunctionComponent<EmailActionProps> = ({
    action,
    setAction,
    actionCompleted,
    setActionCompleted,
    disabled,
    authenticatedUser,
    description,
    cardClassName,
    cardBtnClassName,
    cardLinkClassName,
}) => {
    const [showEmailNotificationForm, setShowEmailNotificationForm] = useState(false)
    const toggleEmailNotificationForm: React.FormEventHandler = useCallback(
        event => {
            if (!disabled) {
                setShowEmailNotificationForm(show => !show)
            }
        },
        [disabled]
    )

    const [emailNotificationEnabled, setEmailNotificationEnabled] = useState(action ? action.enabled : true)

    const toggleEmailNotificationEnabled: (enabled: boolean) => void = useCallback(
        enabled => {
            setEmailNotificationEnabled(enabled)
            setAction({
                __typename: 'MonitorEmail',
                id: action?.id ?? '',
                recipients: { nodes: [{ id: authenticatedUser.id }] },
                enabled,
            })
        },
        [action?.id, authenticatedUser.id, setAction]
    )

    const completeForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setShowEmailNotificationForm(false)
            setActionCompleted(true)
            if (!action) {
                // We are creating a new monitor if there are no actions yet.
                // The ID can be empty here, since we'll generate a new ID when we send the creation request.
                setAction({
                    __typename: 'MonitorEmail',
                    id: '',
                    recipients: { nodes: [{ id: authenticatedUser.id }] },
                    enabled: true,
                })
            }
        },
        [action, authenticatedUser.id, setAction, setActionCompleted]
    )

    const cancelForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setShowEmailNotificationForm(false)
        },
        [setShowEmailNotificationForm]
    )

    const clearForm: () => void = useCallback(() => {
        setShowEmailNotificationForm(false)
        setAction(undefined)
        setActionCompleted(false)
    }, [setAction, setActionCompleted])

    const [isTestEmailSent, setIsTestEmailSent] = useState(false)
    const [triggerTestEmailActionRequest, triggerTestEmailResult] = useEventObservable(
        useCallback(
            (click: Observable<React.MouseEvent<HTMLButtonElement>>) =>
                click.pipe(
                    mergeMap(() =>
                        triggerTestEmailAction({
                            namespace: authenticatedUser.id,
                            description,
                            email: {
                                enabled: true,
                                priority: MonitorEmailPriority.NORMAL,
                                recipients: [authenticatedUser.id],
                                header: '',
                            },
                        }).pipe(
                            delay(1000),
                            startWith(LOADING),
                            tap(value => {
                                if (value !== LOADING) {
                                    setIsTestEmailSent(true)
                                }
                            }),
                            catchError(error => [asError(error)])
                        )
                    )
                ),
            [authenticatedUser, description]
        )
    )

    useEffect(() => {
        if ((isTestEmailSent && !description) || !showEmailNotificationForm) {
            setIsTestEmailSent(false)
        }
    }, [isTestEmailSent, description, showEmailNotificationForm])

    const sendTestEmailButtonText =
        triggerTestEmailResult === LOADING
            ? 'Sending email...'
            : isTestEmailSent
            ? 'Test email sent!'
            : 'Send test email'
    const isSendTestEmailButtonDisabled = triggerTestEmailResult === LOADING || isTestEmailSent || !description

    // When the action is completed, the wrapper cannot be a button because we show nested buttons inside it.
    // Use a div instead. The edit button will still allow keyboard users to activate the form.
    const CollapsedWrapperElement = actionCompleted ? 'div' : Button

    return (
        <>
            {showEmailNotificationForm && (
                <div className={classNames(cardClassName, 'card p-3')}>
                    <div className="font-weight-bold">Send email notifications</div>
                    <span className="text-muted">Deliver email notifications to specified recipients.</span>
                    <div className="form-group mt-4 test-action-form" data-testid="action-form">
                        <label htmlFor="code-monitoring-form-actions-recipients">Recipients</label>
                        <input
                            id="code-monitoring-form-actions-recipients"
                            type="text"
                            className="form-control mb-2"
                            value={`${authenticatedUser.email || ''} (you)`}
                            disabled={true}
                            autoFocus={true}
                            required={true}
                        />
                        <small className="text-muted">
                            Code monitors are currently limited to sending emails to your primary email address.
                        </small>
                    </div>
                    <div className="flex mt-1">
                        <Button
                            className={classNames(
                                'mr-2',
                                isSendTestEmailButtonDisabled ? 'btn-secondary' : 'btn-outline-secondary'
                            )}
                            disabled={isSendTestEmailButtonDisabled}
                            onClick={triggerTestEmailActionRequest}
                            size="sm"
                        >
                            {sendTestEmailButtonText}
                        </Button>
                        {isTestEmailSent && triggerTestEmailResult !== LOADING && (
                            <Button className="p-0" onClick={triggerTestEmailActionRequest} variant="link" size="sm">
                                Send again
                            </Button>
                        )}
                        {!description && (
                            <div className={classNames('mt-2', styles.testActionError)}>
                                Please provide a name for the code monitor before sending a test
                            </div>
                        )}
                        {isErrorLike(triggerTestEmailResult) && (
                            <div className={classNames('mt-2', styles.testActionError)}>
                                {triggerTestEmailResult.message}
                            </div>
                        )}
                    </div>
                    <div className="d-flex align-items-center my-4">
                        <div>
                            <Toggle
                                title="Enabled"
                                value={emailNotificationEnabled}
                                onToggle={toggleEmailNotificationEnabled}
                                className="mr-2"
                                aria-labelledby="code-monitoring-form-actions-enable-toggle"
                            />
                        </div>
                        <span id="code-monitoring-form-actions-enable-toggle">
                            {emailNotificationEnabled ? 'Enabled' : 'Disabled'}
                        </span>
                    </div>
                    <div className="d-flex justify-content-between">
                        <div>
                            <Button
                                type="submit"
                                data-testid="submit-action"
                                className="mr-1 test-submit-action"
                                onClick={completeForm}
                                onSubmit={completeForm}
                                variant="secondary"
                            >
                                Continue
                            </Button>
                            <Button onClick={cancelForm} outline={true} variant="secondary">
                                Cancel
                            </Button>
                        </div>
                        {action && (
                            <Button onClick={clearForm} outline={true} variant="danger">
                                Delete
                            </Button>
                        )}
                    </div>
                </div>
            )}
            {!showEmailNotificationForm && (
                <CollapsedWrapperElement
                    data-testid="form-action-toggle-email-notification"
                    className={classNames('card test-action-button', cardBtnClassName, disabled && 'disabled')}
                    disabled={disabled}
                    aria-label="Edit action: Send email notifications"
                    onClick={toggleEmailNotificationForm}
                >
                    <div className="d-flex justify-content-between align-items-center w-100">
                        <div>
                            <div
                                className={classNames(
                                    'font-weight-bold',
                                    !actionCompleted && classNames(cardLinkClassName, 'btn-link')
                                )}
                            >
                                Send email notifications
                            </div>
                            {actionCompleted ? (
                                <span className="text-muted" data-testid="existing-action-email">
                                    {authenticatedUser.email}
                                </span>
                            ) : (
                                <span className="text-muted">Deliver email notifications to specified recipients.</span>
                            )}
                        </div>
                        {actionCompleted && (
                            <div className="d-flex align-items-center">
                                <div>
                                    <Toggle
                                        title="Enabled"
                                        value={emailNotificationEnabled}
                                        onToggle={toggleEmailNotificationEnabled}
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
