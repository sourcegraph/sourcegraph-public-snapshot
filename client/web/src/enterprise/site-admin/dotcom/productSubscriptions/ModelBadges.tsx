import React from 'react'

import { Badge } from '@sourcegraph/wildcard'

export interface ModelBadgesProps {
    models: string[]
    mode: 'completions' | 'embeddings'
}

export const ModelBadges: React.FunctionComponent<ModelBadgesProps> = ({ models, mode }) => (
    <>
        {models.map(model => (
            <Badge variant={modelBadgeVariant(model, mode)} className="mr-1" key={model}>
                {model}
            </Badge>
        ))}
    </>
)

function modelBadgeVariant(model: string, mode: 'completions' | 'embeddings'): 'secondary' | 'danger' {
    if (mode === 'completions') {
        switch (model) {
            // See here: https://console.anthropic.com/docs/api/reference
            // for currently available Anthropic models. Note that we also need to
            // allow list the models on the Cody Gateway side.
            case 'claude-v1':
            case 'claude-v1.0':
            case 'claude-v1.2':
            case 'claude-v1.3':
            case 'claude-instant-v1':
            case 'claude-instant-v1.0':
            // See here: https://platform.openai.com/docs/models/model-endpoint-compatibility
            // for currently available Anthropic models. Note that we also need to
            // allow list the models on the Cody Gateway side.
            case 'gpt-4':
            case 'gpt-3.5-turbo':
                return 'secondary'
            default:
                return 'danger'
        }
    }
    switch (model) {
        case 'openai/text-embedding-ada-002':
            return 'secondary'
        default:
            return 'danger'
    }
}
