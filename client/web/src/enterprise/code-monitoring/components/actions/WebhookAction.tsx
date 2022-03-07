import { gql, useMutation } from '@apollo/client'
import classNames from 'classnames'
import { noop } from 'lodash'
import React, { useCallback, useMemo, useState } from 'react'

import { Alert, Button, Input, ProductStatusBadge } from '@sourcegraph/wildcard'

import { SendTestWebhookResult, SendTestWebhookVariables } from '../../../../graphql-operations'
import { ActionProps } from '../FormActionArea'
import styles from '../FormActionArea.module.scss'

import { ActionEditor } from './ActionEditor'

export const SEND_TEST_WEBHOOK = gql`
    mutation SendTestWebhook($namespace: ID!, $description: String!, $webhook: MonitorWebhookInput!) {
        triggerTestWebhookAction(namespace: $namespace, description: $description, webhook: $webhook) {
            alwaysNil
        }
    }
`

export const WebhookAction: React.FunctionComponent<ActionProps> = ({
    action,
    setAction,
    disabled,
    monitorName,
    authenticatedUser,
    _testStartOpen,
}) => {
    const [webhookEnabled, setWebhookEnabled] = useState(action ? action.enabled : true)

    const toggleWebhookEnabled: (enabled: boolean) => void = useCallback(
        enabled => {
            setWebhookEnabled(enabled)

            if (action) {
                setAction({ ...action, enabled })
            }
        },
        [action, setAction]
    )

    const [url, setUrl] = useState(action && action.__typename === 'MonitorWebhook' ? action.url : '')
    const urlIsValid = useMemo(() => !!url.match(/^https?:\/\//), [url])

    const onSubmit: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setAction({
                __typename: 'MonitorWebhook',
                id: action ? action.id : '',
                url,
                enabled: webhookEnabled,
                includeResults: false,
            })
        },
        [action, setAction, url, webhookEnabled]
    )

    const onDelete: React.FormEventHandler = useCallback(() => {
        setAction(undefined)
    }, [setAction])

    const [sendTestMessage, { loading, error, called }] = useMutation<SendTestWebhookResult, SendTestWebhookVariables>(
        SEND_TEST_WEBHOOK
    )
    const isSendTestButtonDisabled = loading || !monitorName || !url || (called && !error)

    const onSendTestMessage = useCallback(() => {
        sendTestMessage({
            variables: {
                namespace: authenticatedUser.id,
                description: monitorName,
                webhook: { url, enabled: true, includeResults: false },
            },
        }).catch(noop) // Ignore errors, they will be handled with the error state from useMutation
    }, [authenticatedUser.id, monitorName, sendTestMessage, url])

    const sendTestEmailButtonText = loading
        ? 'Calling webhook...'
        : called && !error
        ? 'Test call completed!'
        : 'Call webhook with test payload'

    return (
        <ActionEditor
            title={
                <div className="d-flex align-items-center">
                    Call a webhook <ProductStatusBadge className="ml-1" status="experimental" />{' '}
                </div>
            }
            label="Call a webhook"
            subtitle="Calls the specified URL with a JSON payload."
            idName="webhook"
            disabled={disabled}
            completed={!!action}
            completedSubtitle="The webhook at the specified URL will be called."
            actionEnabled={webhookEnabled}
            toggleActionEnabled={toggleWebhookEnabled}
            canSubmit={!!urlIsValid}
            onSubmit={onSubmit}
            onCancel={() => {}}
            canDelete={!!action}
            onDelete={onDelete}
            _testStartOpen={_testStartOpen}
        >
            <Alert variant="info" className="mt-4">
                The specified webhook URL will be called with a JSON payload. The format of this JSON payload is still
                being modified. Once it is decided on, documentation will be available.
            </Alert>
            <div className="form-group">
                <label htmlFor="code-monitor-webhook-url">Webhook URL</label>
                <Input
                    id="code-monitor-webhook-url"
                    type="url"
                    className="mb-2"
                    data-testid="webhook-url"
                    required={true}
                    onChange={event => {
                        setUrl(event.target.value)
                    }}
                    value={url}
                    autoFocus={true}
                    spellCheck={false}
                    status={urlIsValid ? 'valid' : url ? 'error' : undefined /* Don't show error state when empty */}
                    message={!urlIsValid && url && 'Enter a valid webhook URL.'}
                />
            </div>
            <div className="flex mt-1">
                <Button
                    className="mr-2"
                    variant="secondary"
                    outline={!isSendTestButtonDisabled}
                    disabled={isSendTestButtonDisabled}
                    onClick={onSendTestMessage}
                    size="sm"
                    data-testid="send-test-webhook"
                >
                    {sendTestEmailButtonText}
                </Button>
                {called && !error && !loading && monitorName && url && (
                    <Button
                        className="p-0"
                        onClick={onSendTestMessage}
                        variant="link"
                        size="sm"
                        data-testid="send-test-webhook-again"
                    >
                        Test again
                    </Button>
                )}
                {!monitorName && (
                    <small className={classNames('mt-2 form-text', styles.testActionError)}>
                        Please provide a name for the code monitor before making a test call
                    </small>
                )}
                {!url && (
                    <small className={classNames('mt-2 form-text', styles.testActionError)}>
                        Please provide a webhook URL before making a test call
                    </small>
                )}
                {error && (
                    <small
                        className={classNames('mt-2 form-text', styles.testActionError)}
                        data-testid="test-webhook-error"
                    >
                        {error.message}
                    </small>
                )}
            </div>
        </ActionEditor>
    )
}
