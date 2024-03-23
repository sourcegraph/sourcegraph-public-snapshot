import React from 'react'

import { Badge } from '@sourcegraph/wildcard'

export interface ModelBadgesProps {
    models: string[]
    mode: 'completions' | 'embeddings'
}

export const ModelBadges: React.FunctionComponent<ModelBadgesProps> = ({ models, mode }) => (
    <>
        {models.map(model => (
            <Badge className="mr-1" key={model} {...modelBadgeProps(model, mode)}>
                {model}
            </Badge>
        ))}
    </>
)

function modelBadgeProps(model: string, mode: 'completions' | 'embeddings'): {variant: 'secondary' | 'danger' | 'warning', tooltip?: string} {
        const ok = {variant: 'secondary'} as const
    // Models also need to be allow-listed on the Cody Gateway side.
    // Anthropic: https://console.anthropic.com/docs/api/reference
    // OpenAI: https://platform.openai.com/docs/models/model-endpoint-compatibility
    switch ([mode, model]) {
        // Anthropic
        case ['completions', 'anthropic/claude-v1']:
        case ['completions', 'anthropic/claude-v1-100k']:
        case ['completions', 'anthropic/claude-2']:
        case ['completions', 'anthropic/claude-instant-v1']:
        case ['completions', 'anthropic/claude-instant-1']:
        case ['completions', 'anthropic/claude-instant-v1-100k']:
            return {variant: 'warning', tooltip: 'Anthropic models should be specified with an explicit minor version.'}
        case ['completions', 'anthropic/claude-v1.0']:
        case ['completions', 'anthropic/claude-v1.2']:
        case ['completions', 'anthropic/claude-v1.3']:
        case ['completions', 'anthropic/claude-v1.3-100k']:
        case ['completions', 'anthropic/claude-2.0']:
        case ['completions', 'anthropic/claude-2.1']:
        case ['completions', 'anthropic/claude-instant-v1.0']:
        case ['completions', 'anthropic/claude-instant-v1.1']:
        case ['completions', 'anthropic/claude-instant-v1.1-100k']:
        case ['completions', 'anthropic/claude-instant-v1.2']:
        case ['completions', 'anthropic/claude-instant-1.2']:
        case ['completions', 'anthropic/claude-3-sonnet-20240229']:
        case ['completions', 'anthropic/claude-3-opus-20240229']:
        case ['completions', 'anthropic/claude-3-haiku-20240307']:
            return ok;
        // OpenAI
        case ['completions', 'openai/gpt-4']:
        case ['completions', 'openai/gpt-3.5-turbo']:
        case ['completions', 'openai/gpt-4-1106-preview']:
        case ['completions', 'openai/gpt-4-turbo-preview']:
        case ['embeddings', 'openai/text-embedding-ada-002']:
            return ok;
        // Virtual models that are translated by Cody Gateway and allow access to all StarCoder
        // models hosted for us by Fireworks.
        case ['completions', 'fireworks/starcoder']:
            return ok;
        // Bespoke alternative models hosted for us by Fireworks. These are also allowed on the
        // Cody Gateway side
        case ['completions', 'fireworks/accounts/fireworks/models/llama-v2-7b-code']:
        case ['completions', 'fireworks/accounts/fireworks/models/llama-v2-13b-code']:
        case ['completions', 'fireworks/accounts/fireworks/models/llama-v2-13b-code-instruct']:
        case ['completions', 'fireworks/accounts/fireworks/models/llama-v2-34b-code-instruct']:
        case ['completions', 'fireworks/accounts/fireworks/models/mistral-7b-instruct-4k']:
        case ['completions', 'fireworks/accounts/fireworks/models/mixtral-8x7b-instruct']:
            return ok;
        default:
            return {variant: 'danger'}
    }
}
