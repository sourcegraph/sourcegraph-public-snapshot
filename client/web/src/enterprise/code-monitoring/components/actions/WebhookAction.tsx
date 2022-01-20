import React from 'react'

import { ProductStatusBadge } from '@sourcegraph/wildcard'

import { ActionEditor } from './ActionEditor'

export const WebhookAction: React.FunctionComponent<{}> = () => (
    <ActionEditor
        title={
            <div className="d-flex align-items-center">
                Call a webhook <ProductStatusBadge className="ml-1" status="experimental" />{' '}
            </div>
        }
        subtitle="Calls the specified URL with a JSON payload."
        disabled={false}
        completed={false}
        completedSubtitle="Action completed"
        actionEnabled={true}
        toggleActionEnabled={() => {}}
        onSubmit={() => {}}
        onCancel={() => {}}
        canDelete={false}
        onDelete={() => {}}
    >
        Coming soon
    </ActionEditor>
)
