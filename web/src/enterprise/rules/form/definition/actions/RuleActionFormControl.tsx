import React from 'react'
import { RULE_ACTION_TYPES, RuleActionTypeComponentContext, GenericRuleAction } from '.'

interface Props extends RuleActionTypeComponentContext<GenericRuleAction> {
    listActions?: React.ReactFragment

    className?: string
}

/**
 * A form control for a single action associated with a rule.
 */
export const RuleActionFormControl: React.FunctionComponent<Props> = ({
    value,
    onChange,
    listActions,
    className = '',
}) => {
    const ruleActionType = RULE_ACTION_TYPES.find(({ id }) => value.type === id)
    if (!ruleActionType) {
        return (
            <div className={`alert alert-danger ${className}`}>
                Unknown rule action type: <code>{value.type}</code>
            </div>
        )
    }
    const { title, icon: Icon, renderForm: RenderForm } = ruleActionType
    return (
        <div className={className}>
            <h5 className="d-flex align-items-center">
                {Icon && <Icon className="icon-inline mr-1" />} {title}
                <div className="flex-1" />
                {listActions}
            </h5>
            {/* Because we checked that value.type === ruleActionType.id above: */
            /* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
            <RenderForm value={value as any} onChange={onChange} />
        </div>
    )
}
