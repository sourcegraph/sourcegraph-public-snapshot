import React, { useState, useCallback } from 'react'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { RULE_ACTION_TYPES, GenericRuleAction } from '.'
import { AddRuleActionDropdownButton } from './AddRuleActionDropdownButton'
import CloseIcon from 'mdi-react/CloseIcon'
import { RuleActionFormControl } from './RuleActionFormControl'

interface Props extends ExtensionsControllerProps {
    className?: string
}

/**
 * A form control for a rule's actions.
 */
export const RuleActionsFormControl: React.FunctionComponent<Props> = ({ className = '' }) => {
    const [ruleActions, setRuleActions] = useState<GenericRuleAction[]>([])
    const onRuleActionChange = useCallback(
        (index: number, value: GenericRuleAction) =>
            setRuleActions([...ruleActions.slice(0, index), value, ...ruleActions.slice(index + 1)]),
        [ruleActions]
    )
    const onNewRuleActionTypeSelect = useCallback(
        (ruleActionInitialValue: GenericRuleAction) => {
            setRuleActions([...ruleActions, ruleActionInitialValue])
        },
        [ruleActions]
    )
    const onRuleActionRemove = useCallback(
        (index: number) => {
            setRuleActions([...ruleActions.slice(0, index), ...ruleActions.slice(index + 1)])
        },
        [ruleActions]
    )
    return (
        <div className={`rule-actions-form-control ${className}`}>
            <AddRuleActionDropdownButton
                items={RULE_ACTION_TYPES}
                onSelect={onNewRuleActionTypeSelect}
                className="mb-2 btn-primary"
            />
            {ruleActions.map((ruleAction, i) => (
                <RuleActionFormControl
                    // eslint-disable-next-line react/no-array-index-key
                    key={i}
                    value={ruleAction}
                    // eslint-disable-next-line react/jsx-no-bind
                    onChange={value => onRuleActionChange(i, value)}
                    listActions={
                        <button
                            type="button"
                            className="btn btn-link text-decoration-none"
                            onClick={() => onRuleActionRemove(i)}
                        >
                            <CloseIcon className="icon-inline" /> Remove
                        </button>
                    }
                />
            ))}
        </div>
    )
}
