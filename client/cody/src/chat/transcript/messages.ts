import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'

export interface ChatMessage extends Message {
    displayText: string
    timestamp: string
    contextFiles?: string[]
}

export interface InteractionMessage extends Message {
    displayText: string
    timestamp: string
    prefix?: string
}
