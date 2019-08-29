import React, { useCallback } from 'react'
import { RuleActionTypeComponentContext, RuleActionType } from '..'
import SlackIcon from 'mdi-react/SlackIcon'

interface SlackRuleAction {
    type: 'slack'
    message: string
    target: string
    targetReviewers: boolean
    remind: boolean
}

// eslint-disable-next-line no-template-curly-in-string
const MESSAGE_LINK = '<${campaign_url}|campaign ${campaign_number}: ${campaign_name}>' // TODO!(sqs): there is no such thing as a "campaign number"

const SlackRuleActionFormControl: React.FunctionComponent<RuleActionTypeComponentContext<SlackRuleAction>> = ({
    value,
    onChange,
}) => {
    const onMessageChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        e => {
            onChange({ ...value, message: e.currentTarget.value })
        },
        [onChange, value]
    )
    const onTargetChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            onChange({ ...value, target: e.currentTarget.value })
        },
        [onChange, value]
    )
    const onTargetReviewersChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            onChange({ ...value, targetReviewers: e.currentTarget.checked })
        },
        [onChange, value]
    )
    const onRemindChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            onChange({ ...value, remind: e.currentTarget.checked })
        },
        [onChange, value]
    )
    return (
        <>
            <div className="form-group">
                <label>Preview</label>
                <div className="rounded border bg-white text-dark p-3 mb-3 d-flex align-items-start">
                    <div className="mr-4 text-muted">12:34</div>
                    {/* TODO!(sqs) hacky interpolation */}
                    <div>
                        <strong className="text-info mr-2">@sourcegraph</strong>{' '}
                        {value.message.replace(MESSAGE_LINK, '')}{' '}
                        <a href="#" className="text-primary">
                            campaign #123: Deprecate react-router-dom (npm) {/* TODO!(sqs): dummy */}
                        </a>
                    </div>
                </div>
            </div>
            <hr className="my-3" />
            <div className="form-group">
                <label htmlFor="rule-action-form-control__message">Message</label>
                <textarea
                    className="form-control"
                    id="rule-action-form-control__message"
                    value={value.message}
                    onChange={onMessageChange}
                    rows={4}
                    required={true}
                />
                <small id="rule-action-form-control__description-help" className="form-text text-muted">
                    Template variables:{' '}
                    <code data-tooltip="The campaign name">
                        {'${campaign_name}'} {/* eslint-disable-line no-template-curly-in-string */}
                    </code>{' '}
                    &nbsp;{' '}
                    <code data-tooltip="The campaign number (example: #49)">
                        {'${campaign_number}'} {/* eslint-disable-line no-template-curly-in-string */}
                    </code>{' '}
                    &nbsp;{' '}
                    <code data-tooltip="The URL to the campaign's page on Sourcegraph">
                        {'${campaign_url}'} {/* eslint-disable-line no-template-curly-in-string */}
                    </code>
                </small>
            </div>
            <div className="form-group">
                <label htmlFor="rule-action-form-control__target">Recipients</label>
                <input
                    type="text"
                    className="form-control"
                    id="rule-action-form-control__target"
                    aria-describedby="rule-action-form-control__target-help"
                    placeholder="@alice, #dev-frontend, @bob"
                    value={value.target}
                    onChange={onTargetChange}
                />
                <small className="form-help text-muted" id="rule-action-form-control__target-help">
                    Supports multiple @usernames and #channels (comma-separated)
                </small>
            </div>
            <div className="form-check mb-3">
                <input
                    className="form-check-input"
                    type="checkbox"
                    id="rule-action-form-control__targetReviewers"
                    checked={value.targetReviewers}
                    onChange={onTargetReviewersChange}
                />
                <label className="form-check-label" htmlFor="rule-action-form-control__targetReviewers">
                    Also send to changeset reviewers
                </label>
            </div>
            <div className="form-check mb-3">
                <input
                    className="form-check-input"
                    type="checkbox"
                    id="rule-action-form-control__remind"
                    checked={value.remind}
                    onChange={onRemindChange}
                />
                <label className="form-check-label" htmlFor="rule-action-form-control__remind">
                    Remind recipients every 48 hours (if action is needed)
                </label>
            </div>
        </>
    )
}

export const SlackRuleActionType: RuleActionType<'slack', SlackRuleAction> = {
    id: 'slack',
    title: 'Slack',
    icon: SlackIcon,
    renderForm: SlackRuleActionFormControl,
    initialValue: {
        type: 'slack',
        message: '🎗️ Your action is needed on ' + MESSAGE_LINK,
        target: '',
        targetReviewers: false,
        remind: false,
    },
}
