import React from 'react'

import { Badge } from '@sourcegraph/wildcard'

export const ModelBadges: React.FunctionComponent<{ models: string[] }> = ({ models }) => (
    <>
        {models.map(model => (
            <Badge variant={modelBadgeVariant(model)} className="mr-1" key={model}>
                {model}
            </Badge>
        ))}
    </>
)

function modelBadgeVariant(model: string): 'secondary' | 'danger' {
    switch (model) {
        case 'claude-v1':
        case 'claude-v1.0':
        case 'claude-v1.2':
        case 'claude-v1.3':
        case 'claude-instant-v1':
        case 'claude-instant-v1.0':
        case 'gpt-4':
        case 'gpt-3.5-turbo':
            return 'secondary'
        default:
            return 'danger'
    }
}
