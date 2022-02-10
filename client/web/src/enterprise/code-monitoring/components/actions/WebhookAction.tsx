import React, { useCallback, useState } from 'react'

import { Alert, Button, ProductStatusBadge } from '@sourcegraph/wildcard'

import { ActionProps } from '../FormActionArea'

import { ActionEditor } from './ActionEditor'

export const WebhookAction: React.FunctionComponent<ActionProps> = ({
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

    const [url, setUrl] = useState(action && action.__typename === 'MonitorWebhook' ? action.url : '')

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
            canSubmit={!!url}
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
                <input
                    id="code-monitor-webhook-url"
                    type="url"
                    className="form-control mb-2"
                    data-testid="webhook-url"
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
