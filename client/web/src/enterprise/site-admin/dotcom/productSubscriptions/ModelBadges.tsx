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
            case 'anthropic/claude-v1':
            case 'anthropic/claude-v1-100k':
            case 'anthropic/claude-v1.0':
            case 'anthropic/claude-v1.2':
            case 'anthropic/claude-v1.3':
            case 'anthropic/claude-v1.3-100k':
            case 'anthropic/claude-2':
            case 'anthropic/claude-instant-v1':
            case 'anthropic/claude-instant-1':
            case 'anthropic/claude-instant-v1-100k':
            case 'anthropic/claude-instant-v1.0':
            case 'anthropic/claude-instant-v1.1':
            case 'anthropic/claude-instant-v1.1-100k':
            case 'anthropic/claude-instant-v1.2':
            case 'anthropic/claude-instant-1.2-cyan':
            // See here: https://platform.openai.com/docs/models/model-endpoint-compatibility
            // for currently available Anthropic models. Note that we also need to
            // allow list the models on the Cody Gateway side.
            case 'openai/gpt-4':
            case 'openai/gpt-3.5-turbo':
            case 'openai/gpt-4-1106-preview':
            // Bespoken models hosted for us by Fireworks. These are also allowed on the Cody
            // Gateway side
            case 'fireworks/accounts/fireworks/models/starcoder-16b-w8a16':
            case 'fireworks/accounts/fireworks/models/starcoder-7b-w8a16':
            case 'fireworks/accounts/fireworks/models/starcoder-3b-w8a16':
            case 'fireworks/accounts/fireworks/models/starcoder-1b-w8a16':
            case 'fireworks/accounts/sourcegraph/models/starcoder-7b':
            case 'fireworks/accounts/sourcegraph/models/starcoder-16b':
            case 'fireworks/accounts/fireworks/models/llama-v2-7b-code':
            case 'fireworks/accounts/fireworks/models/llama-v2-13b-code':
            case 'fireworks/accounts/fireworks/models/llama-v2-13b-code-instruct':
            case 'fireworks/accounts/fireworks/models/llama-v2-34b-code-instruct':
            case 'fireworks/accounts/fireworks/models/mistral-7b-instruct-4k':
            case 'fireworks/accounts/fireworks/models/wizardcoder-15b': {
                return 'secondary'
            }
            default: {
                return 'danger'
            }
        }
    }
    switch (model) {
        case 'openai/text-embedding-ada-002': {
            return 'secondary'
        }
        default: {
            return 'danger'
        }
    }
}
