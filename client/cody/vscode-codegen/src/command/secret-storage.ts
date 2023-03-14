import * as vscode from 'vscode'

export interface SecretStorage {
    get(key: string): Promise<string | undefined>
    store(key: string, value: string): Promise<void>
    delete(key: string): Promise<void>
    onDidChange(callback: (key: string) => void): vscode.Disposable
}

export class VSCodeSecretStorage implements SecretStorage {
    constructor(private secretStorage: vscode.SecretStorage) {}

    async get(key: string): Promise<string | undefined> {
        const value = await this.secretStorage.get(key)
        return value
    }

    async store(key: string, value: string): Promise<void> {
        await this.secretStorage.store(key, value)
    }

    async delete(key: string): Promise<void> {
        await this.secretStorage.delete(key)
    }

    onDidChange(callback: (key: string) => void): vscode.Disposable {
        return this.secretStorage.onDidChange(event => callback(event.key))
    }
}

export class InMemorySecretStorage implements SecretStorage {
    private storage: Map<string, string>
    private callbacks: ((key: string) => void)[]

    constructor() {
        this.storage = new Map<string, string>()
        this.callbacks = []
    }

    async get(key: string): Promise<string | undefined> {
        return this.storage.get(key)
    }

    async store(key: string, value: string): Promise<void> {
        this.storage.set(key, value)

        for (const cb of this.callbacks) {
            cb(key)
        }
    }

    async delete(key: string): Promise<void> {
        this.storage.delete(key)

        for (const cb of this.callbacks) {
            cb(key)
        }
    }

    onDidChange(callback: (key: string) => void): vscode.Disposable {
        this.callbacks.push(callback)

        return new vscode.Disposable(() => {
            const callbackIndex = this.callbacks.indexOf(callback)
            this.callbacks.splice(callbackIndex, 1)
        })
    }
}
