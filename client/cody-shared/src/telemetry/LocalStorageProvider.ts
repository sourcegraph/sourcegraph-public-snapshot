// VS Code Docs https://code.visualstudio.com/api/references/vscode-api#Memento
// A memento represents a storage utility. It can store and retrieve values.
import { Memento } from 'vscode'

export class LocalStorage {
    private ANONYMOUS_USER_ID_KEY = 'sourcegraphAnonymousUid'

    constructor(private storage: Memento) {}

    public getAnonymousUserID(): string | null {
        const anonUserID = this.storage.get(this.ANONYMOUS_USER_ID_KEY, null)
        return anonUserID
    }

    public async setAnonymousUserID(anonUserID: string): Promise<void> {
        try {
            await this.storage.update(this.ANONYMOUS_USER_ID_KEY, anonUserID)
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
