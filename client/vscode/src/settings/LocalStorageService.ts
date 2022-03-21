// VS Code Docs https://code.visualstudio.com/api/references/vscode-api#Memento
// A memento represents a storage utility. It can store and retrieve values.
import { Memento } from 'vscode'

export class LocalStorageService {
    constructor(private storage: Memento) {}

    public getValue(key: string): string {
        return this.storage.get<string>(key, '')
    }

    public async setValue(key: string, value: string): Promise<boolean> {
        try {
            await this.storage.update(key, value)
            return true
        } catch {
            return false
        }
    }

    public instanceVersionWarnings(): string | null {
        const versionNumber = this.storage.get<string>(INSTANCE_VERSION_NUMBER_KEY)
        if (!versionNumber) {
            return 'Cannot determine instance version number'
        }
        if (versionNumber < '3320') {
            return 'Your Sourcegraph instance version is not compatible with this Sourcegraph extension. Please ask your site admin to upgrade the instance to 3.32.0 or above for searches to work with this extension.'
        }
        return null
    }
}

export const SELECTED_SEARCH_CONTEXT_SPEC_KEY = 'selected-search-context-spec'
export const INSTANCE_VERSION_NUMBER_KEY = 'sourcegraphVersionNumber'
export const ANONYMOUS_USER_ID_KEY = 'sourcegraphAnonymousUid'
