import { Message } from '../../sourcegraph-api'

import { TranscriptJSON } from '.'

export interface ChatMessage extends Message {
    displayText?: string
    contextFiles?: string[]
}

export interface InteractionMessage extends Message {
    displayText?: string
    prefix?: string
}

export interface UserLocalHistory {
    chat: ChatHistory
    input: string[]
}

export interface ChatHistory {
    [chatID: string]: TranscriptJSON
}
