import * as vscode from 'vscode'

import { ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { isError } from '@sourcegraph/cody-shared/src/utils'

import {
    AuthStatus,
    DOTCOM_URL,
    LOCAL_APP_URL,
    defaultAuthStatus,
    isLocalApp,
    isLoggedIn as isAuthed,
    unauthenticatedStatus,
    ExtensionMessage,
} from '../chat/protocol'
import { newAuthStatus } from '../chat/utils'
import { logEvent } from '../event-logger'
import { debug } from '../log'

import { AuthMenu, LoginStepInputBox, TokenInputBox } from './AuthMenus'
import { LocalAppDetector } from './LocalAppDetector'
import { LocalStorage } from './LocalStorageProvider'
import { SecretStorage } from './SecretStorageProvider'

export class AuthProvider {
    private endpointHistory: string[] = []

    private appScheme = vscode.env.uriScheme
    private client: SourcegraphGraphQLAPIClient | null = null
    public appDetector: LocalAppDetector

    private authStatus: AuthStatus = defaultAuthStatus
    public webview?: Omit<vscode.Webview, 'postMessage'> & {
        postMessage(message: ExtensionMessage): Thenable<boolean>
    }

    constructor(
        private config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>,
        private secretStorage: SecretStorage,
        private localStorage: LocalStorage
    ) {
        this.authStatus.endpoint = 'init'
        this.loadEndpointHistory()
        this.appDetector = new LocalAppDetector(secretStorage, { onChange: type => this.syncLocalAppState(type) })
    }

    // Sign into the last endpoint the user was signed into
    // if none, try signing in with App URL
    public async init(): Promise<void> {
        await this.appDetector.init()
        const lastEndpoint = this.localStorage?.getEndpoint() || this.config.serverEndpoint
        const token = (await this.secretStorage.get(lastEndpoint || '')) || this.config.accessToken
        debug('AuthProvider:init:lastEndpoint', lastEndpoint)
        const authState = await this.auth(lastEndpoint, token || null)
        if (authState?.isLoggedIn) {
            return
        }
    }

    // Display quickpick to select endpoint to sign in to
    public async signinMenu(type?: 'enterprise' | 'dotcom' | 'token' | 'app', uri?: string): Promise<void> {
        const mode = this.authStatus.isLoggedIn ? 'switch' : 'signin'
        debug('AuthProvider:signinMenu', mode)
        logEvent('CodyVSCodeExtension:login:clicked')
        const item = await AuthMenu(mode, this.endpointHistory)
        if (!item) {
            return
        }
        const menuID = type || item?.id
        switch (menuID) {
            case 'enterprise': {
                const input = await LoginStepInputBox(item.uri, 1, false)
                if (!input?.endpoint) {
                    return
                }
                this.authStatus.endpoint = input.endpoint
                await this.redirectToEndpointLogin(input.endpoint)
                break
            }
            case 'dotcom':
                await this.redirectToEndpointLogin(DOTCOM_URL.href)
                break
            case 'token': {
                const endpoint = uri || item.uri
                const input = await LoginStepInputBox(endpoint, 1, true)
                if (!input?.endpoint || !input?.token) {
                    return
                }
                await this.auth(input.endpoint, input.token)
                break
            }
            case 'app': {
                if (uri) {
                    await this.appAuth(uri)
                }
                break
            }
            default: {
                // Auto log user if token for the selected instance was found in secret
                const selectedEndpoint = item.uri
                const tokenKey = isLocalApp(selectedEndpoint) ? 'SOURCEGRAPH_CODY_APP' : selectedEndpoint
                const token = await this.secretStorage.get(tokenKey)
                const authStatus = await this.auth(selectedEndpoint, token || null)
                this.showIsLoggedIn(authStatus?.authStatus || null)
                if (!authStatus?.isLoggedIn) {
                    const input = await TokenInputBox(item.uri)
                    const authStatusFromToken = await this.auth(selectedEndpoint, input?.token || null)
                    this.showIsLoggedIn(authStatusFromToken?.authStatus || null)
                }
                debug('AuthProvider:signinMenu', mode, selectedEndpoint)
            }
        }
    }

    private showIsLoggedIn(authStatus: AuthStatus | null): void {
        if (!authStatus?.isLoggedIn || !authStatus.endpoint) {
            return
        }
        const endpointName = isLocalApp(authStatus.endpoint) ? 'Cody App' : authStatus.endpoint
        void vscode.window.showInformationMessage(`Signed in to ${endpointName}`)
    }

    public async appAuth(uri?: string): Promise<void> {
        debug('AuthProvider:appAuth:init', '')
        const token = await this.secretStorage.get('SOURCEGRAPH_CODY_APP')
        if (token) {
            const authStatus = await this.auth(LOCAL_APP_URL.href, token)
            if (authStatus?.isLoggedIn) {
                return
            }
        }
        if (!uri) {
            return
        }
        await vscode.env.openExternal(vscode.Uri.parse(uri))
    }

    // Display quickpick to select endpoint to sign out of
    public async signoutMenu(): Promise<void> {
        logEvent('CodyVSCodeExtension:logout:clicked')
        const endpointQuickPickItem = this.authStatus.endpoint ? [this.authStatus.endpoint] : []
        const endpoint = await AuthMenu('signout', endpointQuickPickItem)
        if (!endpoint?.uri) {
            return
        }
        await this.signout(endpoint.uri)
        debug('AuthProvider:signoutMenu', endpoint.uri)
    }

    // Log user out of the selected endpoint (remove token from secret)
    private async signout(endpoint: string): Promise<void> {
        await this.secretStorage.deleteToken(endpoint)
        await this.localStorage.deleteEndpoint()
        await this.auth(endpoint, null)
        this.authStatus.endpoint = ''
        await vscode.commands.executeCommand('setContext', 'cody.activated', false)
    }

    // Create Auth Status
    private async makeAuthStatus(
        config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>
    ): Promise<AuthStatus> {
        const endpoint = config.serverEndpoint
        const token = config.accessToken
        if (!token || !endpoint) {
            return { ...defaultAuthStatus, endpoint }
        }
        // Cache the config and the GraphQL client
        if (this.config !== config || !this.client) {
            this.config = config
            this.client = new SourcegraphGraphQLAPIClient(config)
        }
        // Version is for frontend to check if Cody is not enabled due to unsupported version when siteHasCodyEnabled is false
        const [{ enabled, version }, codyLLMConfiguration] = await Promise.all([
            this.client.isCodyEnabled(),
            this.client.getCodyLLMConfiguration(),
        ])

        const configOverwrites = !isError(codyLLMConfiguration) ? codyLLMConfiguration : undefined

        const isDotComOrApp = this.client.isDotCom() || isLocalApp(endpoint)
        if (!isDotComOrApp) {
            const currentUserID = await this.client.getCurrentUserId()
            const hasVerifiedEmail = false
            return newAuthStatus(
                endpoint,
                isDotComOrApp,
                !isError(currentUserID),
                hasVerifiedEmail,
                enabled,
                version,
                configOverwrites
            )
        }
        const userInfo = await this.client.getCurrentUserIdAndVerifiedEmail()
        const isCodyEnabled = true
        return isError(userInfo)
            ? { ...unauthenticatedStatus, endpoint }
            : newAuthStatus(
                  endpoint,
                  isDotComOrApp,
                  !!userInfo.id,
                  userInfo.hasVerifiedEmail,
                  isCodyEnabled,
                  version,
                  configOverwrites
              )
    }

    public getAuthStatus(): AuthStatus {
        return this.authStatus
    }

    // It processes the authentication steps and stores the login info before sharing the auth status with chatview
    public async auth(
        uri: string,
        token: string | null,
        customHeaders?: {} | null
    ): Promise<{ authStatus: AuthStatus; isLoggedIn: boolean } | null> {
        const endpoint = formatURL(uri) || ''
        const config = {
            serverEndpoint: endpoint,
            accessToken: token,
            customHeaders: customHeaders || this.config.customHeaders,
        }
        const authStatus = await this.makeAuthStatus(config)
        const isLoggedIn = isAuthed(authStatus)
        authStatus.isLoggedIn = isLoggedIn
        await this.storeAuthInfo(endpoint, token)
        await this.syncAuthStatus(authStatus)
        await vscode.commands.executeCommand('setContext', 'cody.activated', isLoggedIn)
        return { authStatus, isLoggedIn }
    }

    // Set auth status and share it with chatview
    private async syncAuthStatus(authStatus: AuthStatus): Promise<void> {
        if (this.authStatus === authStatus) {
            return
        }
        this.authStatus = authStatus
        await this.announceNewAuthStatus()
    }

    public async announceNewAuthStatus(): Promise<void> {
        if (this.authStatus.endpoint === 'init' || !this.webview) {
            return
        }
        await vscode.commands.executeCommand('cody.auth.sync')
    }
    /**
     * Display app state in webview view that is used during Signin flow
     */
    public async syncLocalAppState(type: string): Promise<void> {
        if (this.authStatus.endpoint === 'init' || !this.webview) {
            return
        }
        // Log user into App if user is currently not logged in and has App running
        if (type !== 'app' && !this.authStatus.isLoggedIn) {
            await this.appAuth()
        }
        // Notify webview that app is installed
        await this.webview?.postMessage({ type: 'app-state', isInstalled: true })
    }

    // Register URI Handler (vscode://sourcegraph.cody-ai) for:
    // - Deep linking into VS Code with Cody focused (e.g. from the App setup)
    // - Resolving token sending back from sourcegraph.com and App
    public async tokenCallbackHandler(uri: vscode.Uri, customHeaders: {}): Promise<void> {
        const params = new URLSearchParams(uri.query)
        const isApp = params.get('type') === 'app'
        const token = params.get('code')
        const endpoint = isApp ? LOCAL_APP_URL.href : this.authStatus.endpoint
        if (!token || !endpoint) {
            return
        }
        const authState = await this.auth(endpoint, token, customHeaders)
        if (authState?.isLoggedIn) {
            const successMessage = isApp ? 'Connected to Cody App' : `Signed in to ${endpoint}`
            await vscode.window.showInformationMessage(successMessage)
        }
    }

    // Open callback URL in browser to get token from instance
    public async redirectToEndpointLogin(uri: string): Promise<void> {
        const endpoint = formatURL(uri)
        const isDotComOrApp = uri === LOCAL_APP_URL.href || uri === DOTCOM_URL.href
        if (!endpoint) {
            return
        }
        await fetch(endpoint)
            .then(async res => {
                // Read the string response body
                const version = await res.text()
                if (!isDotComOrApp && version < '5.1.0') {
                    void this.signinMenu('token', uri)
                    return
                }
                const authUri = new URL('/user/settings/tokens/new/callback', endpoint)
                authUri.searchParams.append(
                    'requestFrom',
                    this.appScheme === 'vscode-insiders' ? 'CODY_INSIDERS' : 'CODY'
                )
                this.authStatus.endpoint = endpoint
                // open external link
                void vscode.env.openExternal(vscode.Uri.parse(authUri.href))
            })
            .catch(error => console.error(error))
    }

    // Refresh current endpoint history with the one from local storage
    private loadEndpointHistory(): void {
        this.endpointHistory = this.localStorage.getEndpointHistory() || []
    }

    // Store endpoint in local storage, token in secret storage, and update endpoint history
    private async storeAuthInfo(endpoint: string | null | undefined, token: string | null | undefined): Promise<void> {
        debug('AuthProvider:storeAuthInfo:init', endpoint || '')
        if (!endpoint) {
            return
        }
        await this.localStorage.saveEndpoint(endpoint)
        if (token) {
            await this.secretStorage.storeToken(endpoint, token)
        }
        this.loadEndpointHistory()
        debug('AuthProvider:storeAuthInfo:stored', endpoint || '')
    }
}

function formatURL(uri: string): string | null {
    if (!uri) {
        return null
    }
    // Check if the URI is in the correct URL format
    // Add missing https:// if needed
    if (!uri.startsWith('http')) {
        uri = `https://${uri}`
    }
    try {
        const endpointUri = new URL(uri)
        return endpointUri.href
    } catch {
        console.error('Invalid URL')
    }
    return null
}
