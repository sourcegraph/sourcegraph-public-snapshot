// VS Code Docs https://code.visualstudio.com/api/references/vscode-api#Memento
// A memento represents a storage utility. It can store and retrieve values.
import { Memento } from 'vscode'

// Currently storing:
// sourcegraphSearches: 15 Recent Search History
// sourcegraphFiles: 15 Recent Open Files
// sourcegraphContext: Last selected context
// sourcegraphAnonymousUid: Anonymous User ID Key for Event Logger

export class LocalStorageService {
    constructor(private storage: Memento) {}

    // General Functions
    public getValue(key: string): string {
        return this.storage.get(key, '')
    }

    public async setValue(key: string, value: string): Promise<boolean> {
        try {
            await this.storage.update(key, value)
            return true
        } catch {
            return false
        }
    }

    // For Search History
    public async resetStorage(): Promise<boolean> {
        try {
            await this.storage.update('sourcegraphSearches', undefined)
            await this.storage.update('sourcegraphFiles', undefined)
            await this.storage.update('sourcegraphContext', undefined)
            await this.storage.update('sourcegraphAnonymousUid', undefined)
            return true
        } catch (error) {
            console.error(error)
            return false
        }
    }

    public getLocalSearchHistory(): VsceLocalHistoryProps {
        return {
            searches: this.storage.get<LocalRecentSeachProps[]>('sourcegraphSearches', []),
            files: this.storage.get<string[]>('sourcegraphFiles', []),
            context: this.storage.get<string>('sourcegraphContext', ''),
        }
    }

    public async setLocalSearchHistory(newSearches: LocalRecentSeachProps[]): Promise<boolean> {
        try {
            await this.storage.update('sg-search-history', newSearches)
            return true
        } catch (error) {
            console.error(error)
            return false
        }
    }

    public async setFileHistory(newFile: string[]): Promise<boolean> {
        try {
            await this.storage.update('sourcegraphFiles', newFile)
            return true
        } catch {
            return false
        }
    }
}

// Props
interface LocalRecentSeachProps {
    query: string
    context: string
    caseSensitive: boolean
    patternType: string
    fullQuery: string
}

interface VsceLocalHistoryProps {
    searches: LocalRecentSeachProps[]
    files: string[]
    context: string
}
