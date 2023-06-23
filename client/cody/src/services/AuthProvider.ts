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
} from '../chat/protocol'
import { newAuthStatus } from '../chat/utils'
import { updateConfiguration } from '../configuration'
import { logEvent } from '../event-logger'
import { debug } from '../log'

import { AuthMenu, LoginStepInputBox, TokenInputBox } from './CodyMenus'
import { LocalStorage } from './LocalStorageProvider'
import { SecretStorage } from './SecretStorageProvider'

export class AuthProvider {
    private endpointHistory: string[] = []

    private appScheme = vscode.env.uriScheme
    private client: SourcegraphGraphQLAPIClient | null = null

    private authStatus: AuthStatus = defaultAuthStatus

    constructor(
        private config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>,
        private secretStorage: SecretStorage,
        private localStorage: LocalStorage
    ) {
        this.loadEndpointHistory()
        this.init(localStorage).catch(() => null)
    }

    // Sign into the last endpoint the user was signed into if any
    private async init(localStorage: LocalStorage): Promise<void> {
        const lastEndpoint = localStorage?.getEndpoint()
        if (!lastEndpoint) {
            return
        }
        debug('AuthProvider:init:lastEndpoint', lastEndpoint)
        const token = await this.secretStorage.get(lastEndpoint)
        if (!token) {
            return
        }
        await this.auth(lastEndpoint, token || null)
        debug('AuthProvider:init:tokenFound', lastEndpoint)
    }

    // Display quickpick to select endpoint to sign in to
    public async signinMenu(type?: 'enterprise' | 'dotcom' | 'token', uri?: string): Promise<void> {
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
                await this.redirectToEndpointLogin(false)
                break
            }
            case 'dotcom':
                await this.redirectToEndpointLogin(true)
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
            default: {
                // Auto log user if token for the selected instance was found in secret
                const selectedEndpoint = item.uri
                const token = (await this.secretStorage.get(selectedEndpoint)) || null
                const authState = await this.auth(selectedEndpoint, token)
                if (!authState) {
                    return
                }
                let isLoggedIn = authState.isLoggedIn
                if (!authState.isLoggedIn) {
                    const input = await TokenInputBox(item.uri)
                    isLoggedIn = (await this.auth(selectedEndpoint, input?.token || null))?.isLoggedIn || false
                }
                if (isLoggedIn) {
                    void vscode.window.showInformationMessage(`Signed in to ${selectedEndpoint}`)
                    debug('AuthProvider:signinMenu', mode, selectedEndpoint)
                    return
                }
            }
        }
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
        await updateConfiguration('serverEndpoint', '')
        await this.auth(endpoint, null)
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
        const { enabled, version } = await this.client.isCodyEnabled()
        const isDotComOrApp = this.client.isDotCom() || isLocalApp(endpoint)
        if (!isDotComOrApp) {
            const currentUserID = await this.client.getCurrentUserId()
            const hasVerifiedEmail = false
            return newAuthStatus(endpoint, isDotComOrApp, !isError(currentUserID), hasVerifiedEmail, enabled, version)
        }
        const userInfo = await this.client.getCurrentUserIdAndVerifiedEmail()
        const isCodyEnabled = true
        return isError(userInfo)
            ? { ...unauthenticatedStatus, endpoint }
            : newAuthStatus(endpoint, isDotComOrApp, !!userInfo.id, userInfo.hasVerifiedEmail, isCodyEnabled, version)
    }

    public getAuthStatus(): AuthStatus {
        return this.authStatus
    }

    // It processes the authentication steps and stores the login info before sharing the auth status with chatview
    public async auth(
        uri: string,
        token: string | null,
        customHeaders = {}
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
        await this.syncAuthStatus(authStatus)
        await this.storeAuthInfo(endpoint, token)
        return { authStatus, isLoggedIn }
    }

    // Set auth status and share it with chatview
    private async syncAuthStatus(authStatus: AuthStatus): Promise<void> {
        if (this.authStatus === authStatus) {
            return
        }
        this.authStatus = authStatus
        await vscode.commands.executeCommand('cody.auth.sync', authStatus)
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
        await this.storeAuthInfo(endpoint, token)
        const authState = await this.auth(endpoint, token, customHeaders)
        if (authState?.isLoggedIn) {
            const successMessage = isApp ? 'Connected to Cody App' : `Signed in to ${isLocalApp(endpoint)}`
            await vscode.window.showInformationMessage(successMessage)
        }
    }

    // Open callback URL in browser to get token from instance
    private async redirectToEndpointLogin(isDotCom: boolean): Promise<void> {
        const uri = isDotCom ? DOTCOM_URL.href : this.authStatus.endpoint || ''
        const endpoint = formatURL(uri)
        if (!endpoint) {
            return
        }
        await fetch(endpoint)
            .then(async res => {
                // Read the string response body
                const version = await res.text()
                if (version < '5.1.0') {
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
        await updateConfiguration('serverEndpoint', endpoint)
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
