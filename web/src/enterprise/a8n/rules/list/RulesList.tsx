import H from 'history'
import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { RulesAreaContext } from '../scope/ScopeRulesArea'
import { RulesListItem } from './RulesListItem'

export interface RulesListContext {
    itemClassName?: string
}

interface Props extends Pick<RulesAreaContext, 'rulesURL'>, RulesListContext {
    rules: GQL.IRule[]

    history: H.History
    location: H.Location
}

/**
 * A list of rules.
 */
export const RulesList: React.FunctionComponent<Props> = ({ itemClassName, rules, ...props }) => (
    <div className="rules-list">
        <ul className="list-group mb-0">
            {rules.map(rule => (
                <RulesListItem
                    {...props}
                    key={rule.id}
                    tag="li"
                    rule={rule}
                    className={`list-group-item list-group-item-action ${itemClassName}`}
                />
            ))}
        </ul>
    </div>
)
