import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Collapsible } from '../../../../components/Collapsible'
import { ActionExecutionNode } from './ActionExecutionNode'
import { Link } from '../../../../../../shared/src/components/Link'

export interface ActionNodeProps {
    node: Pick<GQL.IAction, 'id' | 'savedSearch' | 'schedule' | 'actionExecutions'>
}

/**
 * An item in the list of actions.
 */
export const ActionNode: React.FunctionComponent<ActionNodeProps> = ({ node }) => (
    <li className="card p-2 mt-2">
        <Collapsible
            wholeTitleClickable={false}
            title={
                <div className="d-flex align-items-center">
                    <Link to={`/campaigns/actions/${node.id}`} className="d-block">
                        Unnamed action
                    </Link>
                    {node.schedule}
                    {node.savedSearch?.description}
                </div>
            }
        >
            <h4>Associated executions</h4>
            <ul className="list-group">
                {node.actionExecutions.nodes.map(actionExecution => (
                    <ActionExecutionNode node={actionExecution} key={actionExecution.id} />
                ))}
            </ul>
        </Collapsible>
    </li>
)
