import classNames from 'classnames'
import React, { useState, useCallback, useEffect } from 'react'
import { Observable } from 'rxjs'
import { delay, startWith, tap, mergeMap, catchError } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { useEventObservable, Button } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { MonitorEmailPriority } from '../../../../graphql-operations'
import { triggerTestEmailAction } from '../../backend'
import { MonitorAction } from '../FormActionArea'
import styles from '../FormActionArea.module.scss'

import { ActionEditor } from './ActionEditor'

const LOADING = 'LOADING' as const

interface EmailActionProps {
    action?: MonitorAction
    setAction: (action?: MonitorAction) => void
    actionCompleted: boolean
    setActionCompleted: (actionCompleted: boolean) => void
    disabled: boolean
    authenticatedUser: AuthenticatedUser
    description: string
}

export const EmailAction: React.FunctionComponent<EmailActionProps> = ({
    action,
    setAction,
    actionCompleted,
    setActionCompleted,
    disabled,
    authenticatedUser,
    description,
}) => {
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

    const clearForm: React.FormEventHandler = useCallback(() => {
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
        if (isTestEmailSent && !description) {
            setIsTestEmailSent(false)
        }
    }, [isTestEmailSent, description])

    const sendTestEmailButtonText =
        triggerTestEmailResult === LOADING
            ? 'Sending email...'
            : isTestEmailSent
            ? 'Test email sent!'
            : 'Send test email'
    const isSendTestEmailButtonDisabled = triggerTestEmailResult === LOADING || isTestEmailSent || !description

    return (
        <ActionEditor
            title="Send email notifications"
            subtitle="Deliver email notifications to specified recipients."
            disabled={disabled}
            completed={actionCompleted}
            completedSubtitle={authenticatedUser.email}
            actionEnabled={emailNotificationEnabled}
            toggleActionEnabled={toggleEmailNotificationEnabled}
            onSubmit={completeForm}
            canDelete={!!action}
            onDelete={clearForm}
        >
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
                    <div className={classNames('mt-2', styles.testActionError)}>{triggerTestEmailResult.message}</div>
                )}
            </div>
        </ActionEditor>
    )
}
