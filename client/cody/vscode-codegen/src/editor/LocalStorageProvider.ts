// VS Code Docs https://code.visualstudio.com/api/references/vscode-api#Memento
// A memento represents a storage utility. It can store and retrieve values.
import { Memento } from 'vscode'

import { ChatHistory } from '../../webviews/utils/types'

export interface UserLocalHistory {
    chat: ChatHistory
    input: string[]
}

export class LocalStorageProvider {
    private KEY_LOCAL_HISTORY = 'cody-local-chatHistory'

    constructor(private storage: Memento) {}

    public getChatHistory(): UserLocalHistory | null {
        const history = this.storage.get(this.KEY_LOCAL_HISTORY, null)
        return history
    }

    public async setChatHistory(history: UserLocalHistory): Promise<void> {
        try {
            await this.storage.update(this.KEY_LOCAL_HISTORY, history)
        } catch (error) {
            console.error(error)
        }
    }

    public async removeChatHistory(history: UserLocalHistory): Promise<void> {
        try {
            await this.storage.update(this.KEY_LOCAL_HISTORY, null)
        } catch (error) {
            console.error(error)
        }
    }

    public get(key: string): string | null {
        return this.storage.get(key, null)
    }

    public async set(key: string, value: string): Promise<void> {
        try {
            await this.storage.update(key, value)
        } catch (error) {
            console.error(error)
        }
    }
}
