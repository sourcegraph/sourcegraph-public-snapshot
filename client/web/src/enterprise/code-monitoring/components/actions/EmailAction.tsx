import classNames from 'classnames'
import React, { useState, useCallback, useEffect } from 'react'
import { Observable } from 'rxjs'
import { tap, catchError, startWith, mergeMap, delay } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { useEventObservable, Button } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { MonitorEmailPriority } from '../../../../graphql-operations'
import { triggerTestEmailAction as _triggerTestEmailAction } from '../../backend'
import { ActionProps } from '../FormActionArea'
import styles from '../FormActionArea.module.scss'

import { ActionEditor } from './ActionEditor'

const LOADING = 'LOADING' as const

export interface EmailActionProps extends ActionProps {
    authenticatedUser: AuthenticatedUser
    triggerTestEmailAction: typeof _triggerTestEmailAction
}

export const EmailAction: React.FunctionComponent<EmailActionProps> = ({
    action,
    setAction,
    disabled,
    authenticatedUser,
    monitorName,
    triggerTestEmailAction,
    _testStartOpen,
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

    const onSubmit: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
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
        [action, authenticatedUser.id, setAction]
    )

    const onDelete: React.FormEventHandler = useCallback(() => {
        setAction(undefined)
    }, [setAction])

    const [isTestEmailSent, setIsTestEmailSent] = useState(false)
    const [triggerTestEmailActionRequest, triggerTestEmailResult] = useEventObservable(
        useCallback(
            (click: Observable<React.MouseEvent<HTMLButtonElement>>) =>
                click.pipe(
                    mergeMap(() =>
                        triggerTestEmailAction({
                            namespace: authenticatedUser.id,
                            description: monitorName,
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
            [authenticatedUser.id, monitorName, triggerTestEmailAction]
        )
    )

    useEffect(() => {
        if (isTestEmailSent && !monitorName) {
            setIsTestEmailSent(false)
        }
    }, [isTestEmailSent, monitorName])

    const sendTestEmailButtonText =
        triggerTestEmailResult === LOADING
            ? 'Sending email...'
            : isTestEmailSent
            ? 'Test email sent!'
            : 'Send test email'
    const isSendTestEmailButtonDisabled = triggerTestEmailResult === LOADING || isTestEmailSent || !monitorName

    return (
        <ActionEditor
            title="Send email notifications"
            label="Send email notifications"
            subtitle="Deliver email notifications to specified recipients."
            idName="email"
            disabled={disabled}
            completed={!!action}
            completedSubtitle={authenticatedUser.email}
            actionEnabled={emailNotificationEnabled}
            toggleActionEnabled={toggleEmailNotificationEnabled}
            onSubmit={onSubmit}
            canDelete={!!action}
            onDelete={onDelete}
            _testStartOpen={_testStartOpen}
        >
            <div className="form-group mt-4 test-action-form-email" data-testid="action-form-email">
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
                    className="mr-2"
                    variant="secondary"
                    outline={!isSendTestEmailButtonDisabled}
                    disabled={isSendTestEmailButtonDisabled}
                    onClick={triggerTestEmailActionRequest}
                    size="sm"
                    data-testid="send-test-email"
                >
                    {sendTestEmailButtonText}
                </Button>
                {isTestEmailSent && triggerTestEmailResult !== LOADING && (
                    <Button
                        className="p-0"
                        onClick={triggerTestEmailActionRequest}
                        variant="link"
                        size="sm"
                        data-testid="send-test-email-again"
                    >
                        Send again
                    </Button>
                )}
                {!monitorName && (
                    <div className={classNames('mt-2', styles.testActionError)}>
                        Please provide a name for the code monitor before sending a test
                    </div>
                )}
                {isErrorLike(triggerTestEmailResult) && (
                    <div className={classNames('mt-2', styles.testActionError)} data-testid="test-email-error">
                        {triggerTestEmailResult.message}
                    </div>
                )}
            </div>
        </ActionEditor>
    )
}
