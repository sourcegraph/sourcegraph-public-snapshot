import React, { useCallback } from 'react'
import WebhookIcon from 'mdi-react/WebhookIcon'
import { RuleActionType, RuleActionTypeComponentContext } from '..'

interface WebhookRuleAction {
    type: 'webhook'
    url: string
}

const WebhookRuleActionFormControl: React.FunctionComponent<RuleActionTypeComponentContext<WebhookRuleAction>> = ({
    value,
    onChange,
}) => {
    const onUrlChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            onChange({ ...value, url: e.currentTarget.value })
        },
        [onChange, value]
    )
    return (
        <div className="form-group">
            <label htmlFor="rule-action-form-control__url">URL</label>
            <input
                type="url"
                className="form-control"
                id="rule-action-form-control__url"
                placeholder="https://example.com/receive-hook"
                value={value.url}
                onChange={onUrlChange}
                required={true}
            />
        </div>
    )
}

export const WebhookRuleActionType: RuleActionType<'webhook', WebhookRuleAction> = {
    id: 'webhook',
    title: 'Webhook',
    icon: WebhookIcon,
    renderForm: WebhookRuleActionFormControl,
    initialValue: {
        type: 'webhook',
        url: '',
    },
}
