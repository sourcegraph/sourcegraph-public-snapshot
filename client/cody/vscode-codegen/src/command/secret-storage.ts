import * as vscode from 'vscode'

export interface SecretStorage {
    get(key: string): Thenable<string | undefined>
    store(key: string, value: string): Thenable<void>
    delete(key: string): Thenable<void>
    onDidChange(callback: (key: string) => void): vscode.Disposable
}

export class VSCodeSecretStorage implements SecretStorage {
    constructor(private secretStorage: vscode.SecretStorage) {}

    get(key: string): Thenable<string | undefined> {
        return this.secretStorage.get(key)
    }

    store(key: string, value: string): Thenable<void> {
        return this.secretStorage.store(key, value)
    }

    delete(key: string): Thenable<void> {
        return this.secretStorage.delete(key)
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

    get(key: string): Thenable<string | undefined> {
        return Promise.resolve(this.storage.get(key))
    }

    store(key: string, value: string): Thenable<void> {
        this.storage.set(key, value)

        for (const cb of this.callbacks) {
            cb(key)
        }

        return Promise.resolve()
    }

    delete(key: string): Thenable<void> {
        this.storage.delete(key)

        for (const cb of this.callbacks) {
            cb(key)
        }

        return Promise.resolve()
    }

    onDidChange(callback: (key: string) => void): vscode.Disposable {
        this.callbacks.push(callback)

        return new vscode.Disposable(() => {
            const callbackIndex = this.callbacks.indexOf(callback)
            this.callbacks.splice(callbackIndex, 1)
        })
    }
}
