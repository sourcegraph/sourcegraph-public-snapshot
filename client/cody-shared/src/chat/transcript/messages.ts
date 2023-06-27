import { ContextFile } from '../../codebase-context/messages'
import { Message } from '../../sourcegraph-api'

import { TranscriptJSON } from '.'

export interface ChatButton {
    label: string
    action: string
}

export interface ChatMessage extends Message {
    displayText?: string
    contextFiles?: ContextFile[]
    buttons?: ChatButton[]
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

export interface OldChatHistory {
    [chatID: string]: ChatMessage[]
}
