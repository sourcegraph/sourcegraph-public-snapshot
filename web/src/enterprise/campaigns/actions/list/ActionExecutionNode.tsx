import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Link } from '../../../../../../shared/src/components/Link'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckboxBlankCircleIcon from 'mdi-react/CheckboxBlankCircleIcon'
import SyncIcon from 'mdi-react/SyncIcon'

export interface ActionNodeProps {
    node: Pick<GQL.IActionExecution, 'id' | 'invokationReason' | 'status' | 'campaignPlan'>
}

/**
 * An item in the list of action executions.
 */
export const ActionExecutionNode: React.FunctionComponent<ActionNodeProps> = ({ node }) => (
    <li className="list-group-item">
        <div className="ml-2 d-flex justify-content-between align-content-center">
            <div className="flex-grow-1">
                <h3 className="mb-1">
                    <Link to={`/campaigns/actions/executions/${node.id}`} className="d-block">
                        {node.id}
                    </Link>
                </h3>
                <p className="mb-0">{node.invokationReason}</p>
            </div>
            <div className="flex-grow-0">
                {node.status.state === GQL.BackgroundProcessState.COMPLETED && (
                    <div className="d-flex justify-content-end">
                        <CheckboxBlankCircleIcon data-tooltip="Execution is running" className="text-success" />
                    </div>
                )}
                {node.status.state === GQL.BackgroundProcessState.PROCESSING && (
                    <div className="d-flex justify-content-end">
                        <SyncIcon data-tooltip="Execution is running" className="text-info" />
                    </div>
                )}
                {node.status.state === GQL.BackgroundProcessState.CANCELED && (
                    <div className="d-flex justify-content-end">
                        <CheckboxBlankCircleIcon data-tooltip="Execution has been canceled" className="text-gray" />
                    </div>
                )}
                {node.status.state === GQL.BackgroundProcessState.ERRORED && (
                    <>
                        <div className="d-flex justify-content-end">
                            <AlertCircleIcon data-tooltip="Execution has failed" className="text-danger" />
                        </div>
                        <button type="button" className="btn btn-sm btn-secondary">
                            Retry
                        </button>
                    </>
                )}
            </div>
        </div>
    </li>
)
