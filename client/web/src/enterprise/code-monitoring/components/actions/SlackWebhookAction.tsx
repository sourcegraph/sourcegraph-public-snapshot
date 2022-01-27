import React, { useCallback, useState } from 'react'

import { Alert, Button, Link, ProductStatusBadge } from '@sourcegraph/wildcard'

import { ActionProps } from '../FormActionArea'

import { ActionEditor } from './ActionEditor'

export const SlackWebhookAction: React.FunctionComponent<ActionProps> = ({
    action,
    setAction,
    disabled,
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

    const onSubmit: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setAction({
                __typename: 'MonitorSlackWebhook',
                id: action ? action.id : '',
                url,
                enabled: webhookEnabled,
            })
        },
        [action, setAction, url, webhookEnabled]
    )

    const onDelete: React.FormEventHandler = useCallback(() => {
        setAction(undefined)
    }, [setAction])

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
            canSubmit={!!url}
            onSubmit={onSubmit}
            onCancel={() => {}}
            canDelete={!!action}
            onDelete={onDelete}
            _testStartOpen={_testStartOpen}
        >
            <Alert variant="info" className="mt-4">
                Go to{' '}
                <Link to="https://api.slack.com/" target="_blank" rel="noopener">
                    Slack
                </Link>{' '}
                to create a webhook URL. If you already have a Slack webhook URL, paste it in the field below.{' '}
                Documentation coming soon. {/* TODO: Add link to documentation once #27161 is resolved */}
            </Alert>
            <div className="form-group">
                <label htmlFor="code-monitor-slack-webhook-url">Webhook URL</label>
                <input
                    id="code-monitor-slack-webhook-url"
                    type="url"
                    className="form-control mb-2"
                    data-testid="slack-webhook-url"
                    required={true}
                    onChange={event => {
                        setUrl(event.target.value)
                    }}
                    value={url}
                    autoFocus={true}
                    spellCheck={false}
                />
            </div>
            <div className="flex mt-1">
                <Button className="mr-2" disabled={true} size="sm" variant="secondary">
                    Send test message (coming soon)
                </Button>
            </div>
        </ActionEditor>
    )
}
