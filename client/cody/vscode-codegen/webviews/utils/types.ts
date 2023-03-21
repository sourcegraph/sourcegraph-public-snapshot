export interface Message {
    speaker: 'human' | 'assistant'
    text: string
}

export interface ChatMessage extends Message {
    displayText: string
    timestamp: string
    contextFiles?: string[]
}

export type View = 'chat' | 'recipes' | 'about' | 'login' | 'settings' | 'debug'
