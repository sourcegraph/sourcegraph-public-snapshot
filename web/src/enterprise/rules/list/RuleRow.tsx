import PencilIcon from 'mdi-react/PencilIcon'
import React, { useCallback, useState } from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { UpdateRuleForm } from './EditRuleForm'
import { RuleDeleteButton } from './RuleDeleteButton'

interface Props extends ExtensionsControllerProps {
    rule: GQL.IRule

    /** Called when the rule is updated. */
    onRuleUpdate: () => void
}

/**
 * A row in the list of rules.
 */
export const RuleRow: React.FunctionComponent<Props> = ({ rule, onRuleUpdate, ...props }) => {
    const [isEditing, setIsEditing] = useState(false)
    const toggleIsEditing = useCallback(() => setIsEditing(!isEditing), [isEditing])

    return isEditing ? (
        <UpdateRuleForm
            rule={{ ...rule, definition: rule.definition.raw }}
            onRuleUpdate={onRuleUpdate}
            onDismiss={toggleIsEditing}
        />
    ) : (
        <div className="d-flex align-items-center flex-wrap">
            <div className="flex-1">{rule.name}</div>
            <p className="mb-0 flex-1">{rule.description}</p>
            <div className="text-right flex-0">
                <button type="button" className="btn btn-link text-decoration-none" onClick={toggleIsEditing}>
                    <PencilIcon className="icon-inline" /> Edit
                </button>
                <RuleDeleteButton {...props} rule={rule} onDelete={onRuleUpdate} />
            </div>
        </div>
    )
}
