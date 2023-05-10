import * as vscode from 'vscode'

export const CODY_ACCESS_TOKEN_SECRET = 'cody.access-token'

export async function getAccessToken(secretStorage: SecretStorage): Promise<string | null> {
    try {
        return (await secretStorage.get(CODY_ACCESS_TOKEN_SECRET)) || null
    } catch (error) {
        await secretStorage.delete(CODY_ACCESS_TOKEN_SECRET)
        void vscode.window.showErrorMessage(`Failed to retreive access token for Cody: ${error}`)
        return null
    }
}

export interface SecretStorage {
    get(key: string): Thenable<string | undefined>
    store(key: string, value: string): Thenable<void>
    delete(key: string): Thenable<void>
    onDidChange(callback: (key: string) => Promise<void>): vscode.Disposable
}

export class VSCodeSecretStorage implements SecretStorage {
    constructor(private secretStorage: vscode.SecretStorage) {}

    public get(key: string): Thenable<string | undefined> {
        return this.secretStorage.get(key)
    }

    public store(key: string, value: string): Thenable<void> {
        return this.secretStorage.store(key, value)
    }

    public delete(key: string): Thenable<void> {
        return this.secretStorage.delete(key)
    }

    public onDidChange(callback: (key: string) => Promise<void>): vscode.Disposable {
        return this.secretStorage.onDidChange(event => callback(event.key))
    }
}

export class InMemorySecretStorage implements SecretStorage {
    private storage: Map<string, string>
    private callbacks: ((key: string) => Promise<void>)[]

    constructor() {
        this.storage = new Map<string, string>()
        this.callbacks = []
    }

    public get(key: string): Thenable<string | undefined> {
        return Promise.resolve(this.storage.get(key))
    }

    public store(key: string, value: string): Thenable<void> {
        this.storage.set(key, value)

        for (const cb of this.callbacks) {
            // eslint-disable-next-line callback-return
            void cb(key)
        }

        return Promise.resolve()
    }

    public delete(key: string): Thenable<void> {
        this.storage.delete(key)

        for (const cb of this.callbacks) {
            // eslint-disable-next-line callback-return
            void cb(key)
        }

        return Promise.resolve()
    }

    public onDidChange(callback: (key: string) => Promise<void>): vscode.Disposable {
        this.callbacks.push(callback)

        return new vscode.Disposable(() => {
            const callbackIndex = this.callbacks.indexOf(callback)
            this.callbacks.splice(callbackIndex, 1)
        })
    }
}
