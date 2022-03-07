import { gql, useMutation } from '@apollo/client'
import classNames from 'classnames'
import { noop } from 'lodash'
import React, { useCallback, useMemo, useState } from 'react'

import { Alert, Button, Input, Link, ProductStatusBadge } from '@sourcegraph/wildcard'

import { SendTestSlackWebhookResult, SendTestSlackWebhookVariables } from '../../../../graphql-operations'
import { ActionProps } from '../FormActionArea'
import styles from '../FormActionArea.module.scss'

import { ActionEditor } from './ActionEditor'

export const SEND_TEST_SLACK_WEBHOOK = gql`
    mutation SendTestSlackWebhook($namespace: ID!, $description: String!, $slackWebhook: MonitorSlackWebhookInput!) {
        triggerTestSlackWebhookAction(namespace: $namespace, description: $description, slackWebhook: $slackWebhook) {
            alwaysNil
        }
    }
`

export const SlackWebhookAction: React.FunctionComponent<ActionProps> = ({
    action,
    setAction,
    disabled,
    authenticatedUser,
    monitorName,
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

    const [url, setUrl] = useState(action && action.__typename === 'MonitorSlackWebhook' ? action.url : '')
    const urlIsValid = useMemo(() => url.startsWith('https://hooks.slack.com/services/'), [url])

    const onSubmit: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setAction({
                __typename: 'MonitorSlackWebhook',
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

    const [sendTestMessage, { loading, error, called }] = useMutation<
        SendTestSlackWebhookResult,
        SendTestSlackWebhookVariables
    >(SEND_TEST_SLACK_WEBHOOK)
    const isSendTestButtonDisabled = loading || (called && !error) || !monitorName || !url

    const onSendTestMessage = useCallback(() => {
        sendTestMessage({
            variables: {
                namespace: authenticatedUser.id,
                description: monitorName,
                slackWebhook: { url, enabled: true, includeResults: false },
            },
        }).catch(noop) // Ignore errors, they will be handled with the error state from useMutation
    }, [authenticatedUser.id, monitorName, sendTestMessage, url])

    const sendTestEmailButtonText = loading
        ? 'Sending message...'
        : called && !error
        ? 'Test message sent!'
        : 'Send test message'

    return (
        <ActionEditor
            title={
                <div className="d-flex align-items-center">
                    Send Slack message to channel <ProductStatusBadge className="ml-1" status="experimental" />{' '}
                </div>
            }
            label="Send Slack message to channel"
            subtitle="Post to a specified Slack channel. Requires webhook configuration."
            idName="slack-webhook"
            disabled={disabled}
            completed={!!action}
            completedSubtitle="Notification will be sent to the specified Slack webhook URL."
            actionEnabled={webhookEnabled}
            toggleActionEnabled={toggleWebhookEnabled}
            canSubmit={urlIsValid}
            onSubmit={onSubmit}
            onCancel={() => {}}
            canDelete={!!action}
            onDelete={onDelete}
            _testStartOpen={_testStartOpen}
        >
            <Alert variant="info" className="mt-4">
                Go to{' '}
                <Link to="https://api.slack.com/apps" target="_blank" rel="noopener">
                    Slack
                </Link>{' '}
                to create a webhook URL.
                <br />
                <Link to="https://docs.sourcegraph.com/code_monitoring/how-tos/slack" target="_blank" rel="noopener">
                    Read more about how to set up Slack webhooks in the docs.
                </Link>
            </Alert>
            <div className="form-group">
                <label htmlFor="code-monitor-slack-webhook-url">Webhook URL</label>
                <Input
                    id="code-monitor-slack-webhook-url"
                    type="url"
                    className="mb-2"
                    data-testid="slack-webhook-url"
                    required={true}
                    onChange={event => {
                        setUrl(event.target.value)
                    }}
                    value={url}
                    autoFocus={true}
                    spellCheck={false}
                    status={urlIsValid ? 'valid' : url ? 'error' : undefined /* Don't show error state when empty */}
                    message={!urlIsValid && url && 'Enter a valid Slack webhook URL.'}
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
                    data-testid="send-test-slack-webhook"
                >
                    {sendTestEmailButtonText}
                </Button>
                {called && !error && !loading && monitorName && url && (
                    <Button
                        className="p-0"
                        onClick={onSendTestMessage}
                        variant="link"
                        size="sm"
                        data-testid="send-test-slack-webhook-again"
                    >
                        Send again
                    </Button>
                )}
                {!monitorName && (
                    <small className={classNames('mt-2 form-text', styles.testActionError)}>
                        Please provide a name for the code monitor before sending a test
                    </small>
                )}
                {!url && (
                    <small className={classNames('mt-2 form-text', styles.testActionError)}>
                        Please provide a webhook URL before sending a test
                    </small>
                )}
                {error && (
                    <small
                        className={classNames('mt-2 form-text', styles.testActionError)}
                        data-testid="test-slack-webhook-error"
                    >
                        {error.message}
                    </small>
                )}
            </div>
        </ActionEditor>
    )
}
