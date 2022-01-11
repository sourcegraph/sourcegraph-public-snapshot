import classNames from 'classnames'
import React, { useState, useCallback, useEffect } from 'react'
import { Observable } from 'rxjs'
import { delay, startWith, tap, mergeMap, catchError } from 'rxjs/operators'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { asError, isErrorLike } from '@sourcegraph/common'
import { useEventObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Button } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { CodeMonitorFields, MonitorEmailPriority } from '../../../graphql-operations'
import { triggerTestEmailAction } from '../backend'

import styles from './FormActionArea.module.scss'

interface ActionAreaProps {
    actions: CodeMonitorFields['actions']
    actionsCompleted: boolean
    setActionsCompleted: (completed: boolean) => void
    disabled: boolean
    authenticatedUser: AuthenticatedUser
    onActionsChange: (action: CodeMonitorFields['actions']) => void
    description: string
    cardClassName?: string
    cardBtnClassName?: string
    cardLinkClassName?: string
}

const LOADING = 'LOADING' as const

/**
 * TODO farhan: this component is built with the assumption that each monitor has exactly one email action.
 * Refactor to accomodate for more than one.
 */
export const FormActionArea: React.FunctionComponent<ActionAreaProps> = ({
    actions,
    actionsCompleted,
    setActionsCompleted,
    disabled,
    authenticatedUser,
    onActionsChange,
    description,
    cardClassName,
    cardBtnClassName,
    cardLinkClassName,
}) => {
    const [showEmailNotificationForm, setShowEmailNotificationForm] = useState(false)
    const toggleEmailNotificationForm: React.FormEventHandler = useCallback(event => {
        event.preventDefault()
        setShowEmailNotificationForm(show => !show)
    }, [])

    const [emailNotificationEnabled, setEmailNotificationEnabled] = useState(
        actions.nodes[0] ? actions.nodes[0].enabled : true
    )

    const toggleEmailNotificationEnabled: (enabled: boolean) => void = useCallback(
        enabled => {
            setEmailNotificationEnabled(enabled)
            onActionsChange({
                // TODO farhan: refactor to accomodate more than one action.
                nodes: [{ id: actions.nodes[0].id, recipients: { nodes: [{ id: authenticatedUser.id }] }, enabled }],
            })
        },
        [authenticatedUser, onActionsChange, actions.nodes]
    )

    const completeForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setShowEmailNotificationForm(false)
            setActionsCompleted(true)
            if (actions.nodes.length === 0) {
                // We are creating a new monitor if there are no actions yet.
                // The ID can be empty here, since we'll generate a new ID when we send the creation request.
                onActionsChange({
                    nodes: [{ id: '', enabled: true, recipients: { nodes: [{ id: authenticatedUser.id }] } }],
                })
            }
        },
        [setActionsCompleted, setShowEmailNotificationForm, actions.nodes.length, authenticatedUser.id, onActionsChange]
    )
    const cancelForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setShowEmailNotificationForm(false)
        },
        [setShowEmailNotificationForm]
    )

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

    return (
        <>
            <h3 className="mb-1">Actions</h3>
            <span className="text-muted">Run any number of actions in response to an event</span>
            {/* This should be its own component when you can add multiple email actions */}
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
                </div>
            )}
            {!showEmailNotificationForm && (
                <Button
                    data-testid="form-action-toggle-email-notification"
                    className={classNames('card test-action-button', cardBtnClassName)}
                    aria-label="Edit action: Send email notifications"
                    disabled={disabled}
                    onClick={toggleEmailNotificationForm}
                >
                    <div className="d-flex justify-content-between align-items-center w-100">
                        <div>
                            <div
                                className={classNames(
                                    'font-weight-bold',
                                    !actionsCompleted && classNames(cardLinkClassName, 'btn-link')
                                )}
                            >
                                Send email notifications
                            </div>
                            {actionsCompleted ? (
                                <span className="text-muted" data-testid="existing-action-email">
                                    {authenticatedUser.email}
                                </span>
                            ) : (
                                <span className="text-muted">Deliver email notifications to specified recipients.</span>
                            )}
                        </div>
                        {actionsCompleted && (
                            <div className="d-flex align-items-center">
                                <div>
                                    <Toggle
                                        title="Enabled"
                                        value={emailNotificationEnabled}
                                        onToggle={toggleEmailNotificationEnabled}
                                        className="mr-3"
                                    />
                                </div>
                                <div className="btn-link">Edit</div>
                            </div>
                        )}
                    </div>
                </Button>
            )}
            <small className="text-muted">
                What other actions would you like to take?{' '}
                <a href="mailto:feedback@sourcegraph.com" target="_blank" rel="noopener">
                    Share feedback.
                </a>
            </small>
        </>
    )
}
