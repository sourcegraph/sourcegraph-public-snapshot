// VS Code Docs https://code.visualstudio.com/api/references/vscode-api#Memento
// A memento represents a storage utility. It can store and retrieve values.
import { Memento } from 'vscode'

import { LocalRecentSeachProps } from './src/webview/contract'

export class LocalStorageService {
    constructor(private storage: Memento) {}

    public getValue(key: string): string[] {
        return this.storage.get<string[]>(key, [])
    }

    public async setValue(key: string, value: string[]): Promise<boolean> {
        try {
            await this.storage.update(key, value)
            return true
        } catch {
            return false
        }
    }

    public getLocalRecentSearch(): LocalRecentSeachProps[] {
        return this.storage.get<LocalRecentSeachProps[]>('recent_searches', [])
    }

    public async setLocalRecentSearch(newSearches: LocalRecentSeachProps[]): Promise<boolean> {
        try {
            await this.storage.update('recent_searches', newSearches)
            return true
        } catch (error) {
            console.log(error)
            return false
        }
    }

    public getFileHistory(): string[] {
        return this.storage.get<string[]>('sg-files-test2', [])
    }

    public async setFileHistory(newFile: string[]): Promise<boolean> {
        try {
            await this.storage.update('sg-files-test2', newFile)
            return true
        } catch {
            return false
        }
    }
}
