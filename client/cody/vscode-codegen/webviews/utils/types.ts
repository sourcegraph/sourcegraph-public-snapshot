export interface Message {
	speaker: 'you' | 'bot'
	text: string
}

export interface ChatMessage extends Omit<Message, 'text'> {
	displayText: string
	timestamp: string
	contextFiles?: string[]
}

export type View = 'chat' | 'recipes' | 'about' | 'login' | 'settings' | 'debug'
