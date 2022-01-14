import React from 'react'

import { ProductStatusBadge } from '@sourcegraph/wildcard'

import { ActionEditor } from './ActionEditor'

export const SlackWebhookAction: React.FunctionComponent<{}> = () => (
    <ActionEditor
        title={
            <div className="d-flex align-items-center">
                Send Slack message to channel <ProductStatusBadge className="ml-1" status="experimental" />{' '}
            </div>
        }
        subtitle="Post to a specified Slack channel. Requires webhook configuration."
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
