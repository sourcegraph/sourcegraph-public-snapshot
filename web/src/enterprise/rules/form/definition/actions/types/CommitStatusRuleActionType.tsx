import React, { useCallback } from 'react'
import { RuleActionType, RuleActionTypeComponentContext } from '..'
import { GitCommitIcon } from '../../../../../../util/octicons'

interface CommitStatusRuleAction {
    type: 'commitStatus'
    refs: string
    infoOnly?: boolean
}

const CommitStatusRuleActionFormControl: React.FunctionComponent<
    RuleActionTypeComponentContext<CommitStatusRuleAction>
> = ({ value, onChange }) => {
    const onRefsChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            onChange({ ...value, refs: e.currentTarget.value })
        },
        [onChange, value]
    )
    const onInfoOnlyChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            onChange({ ...value, infoOnly: e.currentTarget.checked })
        },
        [onChange, value]
    )
    return (
        <>
            <div className="form-group">
                <label htmlFor="rule-action-form-control__refs">Branch</label>
                <input
                    type="text"
                    className="form-control"
                    id="rule-action-form-control__refs"
                    value={value.refs}
                    onChange={onRefsChange}
                />
            </div>
            <div className="form-check mb-3">
                <input
                    className="form-check-input"
                    type="checkbox"
                    id="rule-action-form-control__infoOnly"
                    checked={value.infoOnly}
                    onChange={onInfoOnlyChange}
                />
                <label className="form-check-label" htmlFor="rule-action-form-control__infoOnly">
                    Informational only (never fail build)
                </label>
            </div>
        </>
    )
}

export const CommitStatusRuleActionType: RuleActionType<'commitStatus', CommitStatusRuleAction> = {
    id: 'commitStatus',
    title: 'Commit status',
    icon: GitCommitIcon,
    renderForm: CommitStatusRuleActionFormControl,
    initialValue: {
        type: 'commitStatus',
        refs: 'master', // TODO!(sqs)
        infoOnly: false,
    },
}
