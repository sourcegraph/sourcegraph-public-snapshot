import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'

export interface ContextMessage extends Message {
    fileName?: string
}

export function getContextMessageWithResponse(text: string, fileName: string): ContextMessage[] {
    return [
        { speaker: 'human', text, fileName },
        { speaker: 'assistant', text: 'Ok.' },
    ]
}
