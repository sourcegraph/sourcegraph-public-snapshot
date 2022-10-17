import {
    authentication,
    AuthenticationProvider,
    AuthenticationProviderAuthenticationSessionsChangeEvent,
    AuthenticationSession,
    ConfigurationTarget,
    Disposable,
    Event,
    EventEmitter,
    SecretStorage,
} from 'vscode'

import { observeAuthenticatedUser } from '../../backend/authenticatedUser'
import { endpointSetting } from '../../settings/endpointSetting'
import { readConfiguration } from '../../settings/readConfiguration'

export const scretTokenKey = 'SOURCEGRAPH_TOKEN'
class SourcegraphAuthSession implements AuthenticationSession {
    public readonly account = { id: SourcegraphAuthProvider.id, label: 'Sourcegraph Token' }
    public readonly id = SourcegraphAuthProvider.id
    public readonly scopes = []
    constructor(public readonly accessToken: string) {}
}

export class SourcegraphAuthProvider implements AuthenticationProvider, Disposable {
    public static id = endpointSetting() || 'https://sourcegraph.com'
    private static secretKey = 'SOURCEGRAPH_TOKEN'
    // Kept track of token changes through out the session
    private currentToken: string | undefined
    private initializedDisposable: Disposable | undefined
    private _onDidChangeSessions = new EventEmitter<AuthenticationProviderAuthenticationSessionsChangeEvent>()
    public get onDidChangeSessions(): Event<AuthenticationProviderAuthenticationSessionsChangeEvent> {
        return this._onDidChangeSessions.event
    }

    constructor(private readonly secretStorage: SecretStorage) {
        this.init()
            .then(() => {})
            .catch(() => {})
    }

    public async init(): Promise<void> {
        // Process the token that lives in user configuration
        // Move them to secrets and mark them as REDACTED
        const oldToken = readConfiguration().get<string>('accessToken')
        if (oldToken !== undefined && oldToken !== 'REDACTED') {
            await this.secretStorage.store(SourcegraphAuthProvider.secretKey, oldToken)
            await readConfiguration().update('accessToken', 'REDACTED', ConfigurationTarget.Global)
        }
        // delete existing token stored in secret if it is currently not working
        const authenticatedUser = observeAuthenticatedUser(this.secretStorage)
        if (!authenticatedUser) {
            await this.secretStorage.delete(SourcegraphAuthProvider.secretKey)
        }
    }

    public dispose(): void {
        this.initializedDisposable?.dispose()
    }

    private async ensureInitialized(): Promise<void> {
        await this.init()
        if (this.initializedDisposable === undefined) {
            await this.cacheTokenFromStorage()
            this.initializedDisposable = Disposable.from(
                this.secretStorage.onDidChange(async event => {
                    if (event.key === SourcegraphAuthProvider.secretKey) {
                        await this.checkForUpdates()
                    }
                }),
                authentication.onDidChangeSessions(async event => {
                    if (event.provider.id === SourcegraphAuthProvider.id) {
                        await this.checkForUpdates()
                    }
                })
            )
        }
    }

    // Check if token has been updated across VS Code
    private async checkForUpdates(): Promise<void> {
        const added: AuthenticationSession[] = []
        const removed: AuthenticationSession[] = []
        const changed: AuthenticationSession[] = []
        const previousToken = this.currentToken
        const session = (await this.getSessions())[0]
        if (session?.accessToken && !previousToken) {
            added.push(session)
        } else if (!session?.accessToken && previousToken) {
            removed.push(session)
        } else if (session?.accessToken !== previousToken) {
            changed.push(session)
        } else {
            return
        }
        await this.cacheTokenFromStorage()
        this._onDidChangeSessions.fire({ added, removed, changed })
    }

    // Get token from Storage
    private async cacheTokenFromStorage(): Promise<string | undefined> {
        const token = await this.secretStorage.get(SourcegraphAuthProvider.secretKey)
        this.currentToken = token
        return this.currentToken
    }

    // This is called first when `vscode.authentication.getSessions` is called.
    public async getSessions(_scopes?: string[]): Promise<readonly AuthenticationSession[]> {
        await this.ensureInitialized()
        const token = await this.cacheTokenFromStorage()
        return token ? [new SourcegraphAuthSession(token)] : []
    }

    // This is called after `this.getSessions` and only when `createIfNone` and `forceNewSession` are set to true
    public async createSession(_scopes: string[]): Promise<AuthenticationSession> {
        await this.ensureInitialized()
        // Get token from scret storage
        let token = await this.secretStorage.get(SourcegraphAuthProvider.secretKey)
        if (token) {
            console.log('Successfully logged in to Sourcegraph', token)
            await this.secretStorage.store(SourcegraphAuthProvider.secretKey, token)
        }
        if (!token) {
            token = ''
        }
        return new SourcegraphAuthSession(token)
    }

    // To sign out
    public async removeSession(_sessionId: string): Promise<void> {
        await this.secretStorage.delete(SourcegraphAuthProvider.secretKey)
        console.log('Successfully logged out of Sourcegraph')
    }
}
