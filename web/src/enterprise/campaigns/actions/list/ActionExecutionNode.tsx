import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Link } from '../../../../../../shared/src/components/Link'

export interface ActionNodeProps {
    node: Pick<GQL.IActionExecution, 'id'>
}

/**
 * An item in the list of action executions.
 */
export const ActionExecutionNode: React.FunctionComponent<ActionNodeProps> = ({ node }) => (
    <li className="list-group-item">
        <Link to={`/campaigns/actions/executions/${node.id}`}>{node.id}</Link>
    </li>
)
