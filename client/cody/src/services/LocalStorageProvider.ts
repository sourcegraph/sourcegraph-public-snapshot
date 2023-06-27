// VS Code Docs https://code.visualstudio.com/api/references/vscode-api#Memento
// A memento represents a storage utility. It can store and retrieve values.
import * as uuid from 'uuid'
// import * as vscode from 'vscode'
import { Memento } from 'vscode'

import { UserLocalHistory } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

export class LocalStorage {
    // Bump this on storage changes so we don't handle incorrectly formatted data
    private KEY_LOCAL_HISTORY = 'cody-local-chatHistory-v2'
    private ANONYMOUS_USER_ID_KEY = 'sourcegraphAnonymousUid'
    private LAST_USED_ENDPOINT = 'SOURCEGRAPH_CODY_ENDPOINT'
    private CODY_ENDPOINT_HISTORY = 'SOURCEGRAPH_CODY_ENDPOINT_HISTORY'

    constructor(private storage: Memento) {}

    public getEndpoint(): string | null {
        return this.storage.get<string | null>(this.LAST_USED_ENDPOINT, null)
    }

    public async saveEndpoint(endpoint: string): Promise<void> {
        if (!endpoint) {
            return
        }
        try {
            const uri = new URL(endpoint).href
            await this.storage.update(this.LAST_USED_ENDPOINT, uri)
            await this.addEndpointHistory(uri)
        } catch (error) {
            console.error(error)
        }
    }

    public async deleteEndpoint(): Promise<void> {
        await this.storage.update(this.LAST_USED_ENDPOINT, null)
    }

    public getEndpointHistory(): string[] | null {
        return this.storage.get<string[] | null>(this.CODY_ENDPOINT_HISTORY, null)
    }

    public async deleteEndpointHistory(): Promise<void> {
        await this.storage.update(this.CODY_ENDPOINT_HISTORY, null)
    }

    private async addEndpointHistory(endpoint: string): Promise<void> {
        const history = this.storage.get<string[] | null>(this.CODY_ENDPOINT_HISTORY, null)
        const historySet = new Set(history)
        historySet.delete(endpoint)
        historySet.add(endpoint)
        await this.storage.update(this.CODY_ENDPOINT_HISTORY, [...historySet])
    }

    public getChatHistory(): UserLocalHistory | null {
        const history = this.storage.get<UserLocalHistory | null>(this.KEY_LOCAL_HISTORY, null)
        return history
    }

    public async setChatHistory(history: UserLocalHistory): Promise<void> {
        try {
            await this.storage.update(this.KEY_LOCAL_HISTORY, history)
        } catch (error) {
            console.error(error)
        }
    }

    public async deleteChatHistory(chatID: string): Promise<void> {
        const userHistory = this.getChatHistory()
        if (userHistory) {
            try {
                delete userHistory.chat[chatID]
                await this.storage.update(this.KEY_LOCAL_HISTORY, { ...userHistory })
            } catch (error) {
                console.error(error)
            }
        }
    }

    public async removeChatHistory(): Promise<void> {
        try {
            await this.storage.update(this.KEY_LOCAL_HISTORY, null)
        } catch (error) {
            console.error(error)
        }
    }

    public getAnonymousUserID(): string | null {
        const anonUserID = this.storage.get(this.ANONYMOUS_USER_ID_KEY, null)
        return anonUserID
    }

    public async setAnonymousUserID(): Promise<string | null> {
        let status: string | null = null
        let anonUserID = this.storage.get(this.ANONYMOUS_USER_ID_KEY)
        if (!anonUserID) {
            anonUserID = uuid.v4()
            status = 'installed'
        }
        try {
            await this.storage.update(this.ANONYMOUS_USER_ID_KEY, anonUserID)
        } catch (error) {
            console.error(error)
        }
        return status
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
