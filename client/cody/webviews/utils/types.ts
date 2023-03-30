export interface Message {
    speaker: 'human' | 'assistant'
    text: string
}

export interface ChatMessage extends Message {
    displayText: string
    timestamp: string
    contextFiles?: string[]
}

export interface UserLocalHistory {
    chat: ChatHistory
    input: string[]
}

export interface ChatHistory {
    [chatID: string]: ChatMessage[]
}

export type View = 'chat' | 'recipes' | 'about' | 'login' | 'settings' | 'debug' | 'history'
