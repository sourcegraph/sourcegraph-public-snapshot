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
        // See here: https://console.anthropic.com/docs/api/reference
        // for currently available Anthropic models. Note that we also need to
        // allow list the models on the Cody Gateway side.
        case 'anthropic/claude-v1':
        case 'anthropic/claude-v1.0':
        case 'anthropic/claude-v1.2':
        case 'anthropic/claude-v1.3':
        case 'anthropic/claude-instant-v1':
        case 'anthropic/claude-instant-v1.0':
        // See here: https://platform.openai.com/docs/models/model-endpoint-compatibility
        // for currently available Anthropic models. Note that we also need to
        // allow list the models on the Cody Gateway side.
        case 'openai/gpt-4':
        case 'openai/gpt-3.5-turbo':
            return 'secondary'
        default:
            return 'danger'
    }
}
