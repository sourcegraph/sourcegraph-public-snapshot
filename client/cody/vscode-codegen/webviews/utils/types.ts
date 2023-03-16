export interface Message {
    speaker: 'human' | 'assistant'
    text: string
}

export interface ChatMessage extends Omit<Message, 'text'> {
    displayText: string
    timestamp: string
    contextFiles?: string[]
}

export type View = 'chat' | 'recipes' | 'about' | 'login' | 'settings' | 'debug'
