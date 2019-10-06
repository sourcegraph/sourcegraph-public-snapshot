import React, { useCallback } from 'react'
import EmailIcon from 'mdi-react/EmailIcon'
import { RuleActionType, RuleActionTypeComponentContext } from '..'

interface EmailRuleAction {
    type: 'email'
    subject: string
    to: string
    cc: string
}

const EmailRuleActionFormControl: React.FunctionComponent<RuleActionTypeComponentContext<EmailRuleAction>> = ({
    value,
    onChange,
}) => {
    const onSubjectChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            onChange({ ...value, subject: e.currentTarget.value })
        },
        [onChange, value]
    )
    const onToChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            onChange({ ...value, to: e.currentTarget.value })
        },
        [onChange, value]
    )
    const onCcChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            onChange({ ...value, cc: e.currentTarget.value })
        },
        [onChange, value]
    )
    return (
        <>
            <div className="form-group">
                <label htmlFor="rule-action-form-control__subject">Subject</label>
                <input
                    type="text"
                    className="form-control"
                    id="rule-action-form-control__subject"
                    value={value.subject}
                    onChange={onSubjectChange}
                    required={true}
                />
            </div>
            <div className="form-group">
                <label htmlFor="rule-action-form-control__to">To</label>
                <input
                    type="text"
                    className="form-control"
                    id="rule-action-form-control__to"
                    aria-describedby="rule-action-form-control__to-help"
                    placeholder="alice@example.com, bob@example.com"
                    value={value.to}
                    onChange={onToChange}
                    required={true}
                />
                <small className="form-help text-muted" id="rule-action-form-control__to-help">
                    Separate multiple email addresses with commas
                </small>
            </div>
            <div className="form-group">
                <label htmlFor="rule-action-form-control__cc">Cc</label>
                <input
                    type="text"
                    className="form-control"
                    id="rule-action-form-control__cc"
                    aria-describedby="rule-action-form-control__cc-help"
                    placeholder="alice@example.com, bob@example.com"
                    value={value.cc}
                    onChange={onCcChange}
                />
                <small className="form-help text-muted" id="rule-action-form-control__cc-help">
                    Separate multiple email addresses with commas
                </small>
            </div>
        </>
    )
}

export const EmailRuleActionType: RuleActionType<'email', EmailRuleAction> = {
    id: 'email',
    title: 'Email',
    icon: EmailIcon,
    renderForm: EmailRuleActionFormControl,
    initialValue: {
        type: 'email',
        subject: `🎯 ${'TODO!(sqs): thread.title'}`,
        to: '',
        cc: '',
    },
}
