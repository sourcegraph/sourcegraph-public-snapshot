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
}
