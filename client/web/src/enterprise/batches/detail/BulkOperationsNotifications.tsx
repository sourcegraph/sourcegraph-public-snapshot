import * as H from 'history'
import React from 'react'

import { DismissibleAlert } from '../../../components/DismissibleAlert'
import { ActiveBulkOperationsConnectionFields } from '../../../graphql-operations'

import { BulkOperationNode } from './bulk-operations/BulkOperationNode'

export interface BulkOperationsNotificationsProps {
    location: H.Location
    bulkOperations: ActiveBulkOperationsConnectionFields
}

export const BulkOperationsNotifications: React.FunctionComponent<BulkOperationsNotificationsProps> = ({
    bulkOperations,
    location,
}) => {
    // Don't show the header banners if the bulkoperations tab is open.
    const parameters = new URLSearchParams(location.search)
    if (parameters.get('tab') === 'bulkoperations') {
        return null
    }

    return (
        <>
            {bulkOperations.nodes.map(node => (
                <DismissibleAlert
                    key={node.id}
                    className="alert alert-info"
                    partialStorageKey={`bulkOperation-${node.id}`}
                >
                    <BulkOperationNode node={node} key={node.id} />
                </DismissibleAlert>
            ))}
        </>
    )
}
