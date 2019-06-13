import PencilIcon from 'mdi-react/PencilIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { RuleDeleteButton } from '../components/RuleDeleteButton'

interface Props extends ExtensionsControllerProps {
    rule: GQL.IRule

    /** Called when the rule is updated. */
    onRuleUpdate: () => void
}

/**
 * A row in the list of rules.
 */
export const RuleRow: React.FunctionComponent<Props> = ({ rule, onRuleUpdate, ...props }) => (
    <div className="d-flex align-items-center flex-wrap">
        <div className="flex-1">{rule.name}</div>
        <p className="mb-0 flex-1">{rule.description}</p>
        <div className="text-right flex-0">
            <Link to={rule.url} className="btn btn-link text-decoration-none">
                <PencilIcon className="icon-inline" /> Edit
            </Link>
            <RuleDeleteButton {...props} rule={rule} onDelete={onRuleUpdate} />
        </div>
    </div>
)
