import * as vscode from 'vscode'

export class SecretStorage {
    private token = ''
    private url = 'https://sourcegraph.com'

    constructor(private secretStorage: vscode.SecretStorage) {
        this.init()
            .then(() => {})
            .catch(() => {})
    }

    public async init(): Promise<void> {
        const token = await this.secretStorage.get('SOURCEGRAPH_TOKEN')
        const url = await this.secretStorage.get('SOURCEGRAPH_URL')
        this.token = token || ''
        this.url = url || this.url
    }
    public async storeToken(token?: string): Promise<void> {
        if (token) {
            this.token = token
            await this.secretStorage.store('SOURCEGRAPH_TOKEN', token)
        }
    }
    public async storeSecrets(url: string, token?: string): Promise<void> {
        this.removeEndingSlash(url)
        this.url = url
        this.token = token || ''
        await this.secretStorage.store('SOURCEGRAPH_TOKEN', this.token)
        await this.secretStorage.store('SOURCEGRAPH_URL', this.url)
    }
    public getToken(): string {
        return this.token
    }
    public getURL(): string {
        return this.url
    }
    private removeEndingSlash(uri: string): string {
        if (uri.endsWith('/')) {
            return uri.slice(0, -1)
        }
        return uri
    }
}
